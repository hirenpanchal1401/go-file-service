package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"file-service/proto"
	"file-service/service"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func getPort() string {
	port := os.Getenv("FILE_SERVICE_PORT")
	if port == "" {
		port = "50051"
	}
	// Ensure port starts with ':'
	if port[0] != ':' {
		port = ":" + port
	}
	return port
}

type fileServiceServer struct {
	proto.UnimplementedFileServiceServer
	pdfProcessor   *service.PDFProcessor
	imageProcessor *service.ImageProcessor
}

func NewFileServiceServer(tempDir string) *fileServiceServer {
	return &fileServiceServer{
		pdfProcessor:   service.NewPDFProcessor(tempDir),
		imageProcessor: service.NewImageProcessor(),
	}
}

// CompressImage compresses an image file
func (s *fileServiceServer) CompressImage(ctx context.Context, req *proto.ImageRequest) (*proto.ImageResponse, error) {
	log.Printf("Received CompressImage request: size=%d, mime=%s, quality=%d", len(req.FileData), req.MimeType, req.Quality)

	if len(req.FileData) == 0 {
		return &proto.ImageResponse{
			Status:       "error",
			ErrorMessage: "empty file data",
		}, nil
	}

	originalSize := int64(len(req.FileData))
	quality := int(req.Quality)
	if quality == 0 {
		quality = 85
	}

	compressedData, err := s.imageProcessor.CompressImage(req.FileData, req.MimeType, quality)
	if err != nil {
		log.Printf("Error compressing image: %v", err)
		return &proto.ImageResponse{
			Status:         "error",
			ErrorMessage:   err.Error(),
			OriginalSize:   originalSize,
			CompressedSize: originalSize,
		}, nil
	}

	compressedSize := int64(len(compressedData))
	ratio := service.CalculateCompressionRatio(originalSize, compressedSize)

	log.Printf("Image compressed: %d -> %d bytes (%.2f%% reduction)", originalSize, compressedSize, ratio)

	return &proto.ImageResponse{
		Status:            "success",
		CompressedData:    compressedData,
		OriginalSize:      originalSize,
		CompressedSize:    compressedSize,
		CompressionRatio:  ratio,
	}, nil
}

// CompressPDF compresses a valid PDF
func (s *fileServiceServer) CompressPDF(ctx context.Context, req *proto.PDFRequest) (*proto.PDFResponse, error) {
	log.Printf("Received CompressPDF request: size=%d, operation=%s", len(req.FileData), req.Operation)

	if len(req.FileData) == 0 {
		return &proto.PDFResponse{
			Status:       "error",
			ErrorMessage: "empty file data",
		}, nil
	}

	originalSize := int64(len(req.FileData))

	// Validate PDF first
	if err := s.pdfProcessor.ValidatePDF(req.FileData); err != nil {
		log.Printf("PDF validation failed: %v, attempting repair", err)
		// PDF is corrupted, use RepairPDF instead
		return s.repairAndCompressPDF(req.FileData), nil
	}

	// PDF is valid, just compress
	compressedData, err := s.pdfProcessor.CompressPDF(req.FileData)
	if err != nil {
		log.Printf("Error compressing PDF: %v", err)
		return &proto.PDFResponse{
			Status:         "error",
			ErrorMessage:   err.Error(),
			OriginalSize:   originalSize,
			CompressedSize: originalSize,
		}, nil
	}

	compressedSize := int64(len(compressedData))
	ratio := service.CalculateCompressionRatio(originalSize, compressedSize)

	log.Printf("PDF compressed: %d -> %d bytes (%.2f%% reduction)", originalSize, compressedSize, ratio)

	return &proto.PDFResponse{
		Status:           "success",
		ProcessedData:    compressedData,
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: ratio,
		WasCorrupted:     false,
		WasRepaired:      false,
	}, nil
}

// RepairPDF repairs a corrupted PDF
func (s *fileServiceServer) RepairPDF(ctx context.Context, req *proto.PDFRequest) (*proto.PDFResponse, error) {
	log.Printf("Received RepairPDF request: size=%d", len(req.FileData))

	if len(req.FileData) == 0 {
		return &proto.PDFResponse{
			Status:       "error",
			ErrorMessage: "empty file data",
		}, nil
	}

	return s.repairAndCompressPDF(req.FileData), nil
}

func (s *fileServiceServer) repairAndCompressPDF(fileData []byte) *proto.PDFResponse {
	originalSize := int64(len(fileData))

	// Attempt to repair and compress
	processedData, err := s.pdfProcessor.RepairPDF(fileData)
	if err != nil {
		log.Printf("Error repairing PDF: %v", err)
		return &proto.PDFResponse{
			Status:         "error",
			ErrorMessage:   err.Error(),
			OriginalSize:   originalSize,
			CompressedSize: originalSize,
			WasCorrupted:   true,
			WasRepaired:    false,
		}
	}

	processedSize := int64(len(processedData))
	ratio := service.CalculateCompressionRatio(originalSize, processedSize)

	log.Printf("PDF repaired and compressed: %d -> %d bytes (%.2f%% reduction)", originalSize, processedSize, ratio)

	return &proto.PDFResponse{
		Status:           "success",
		ProcessedData:    processedData,
		OriginalSize:     originalSize,
		CompressedSize:   processedSize,
		CompressionRatio: ratio,
		WasCorrupted:     true,
		WasRepaired:      true,
	}
}

// GetFileStats returns statistics about a file
func (s *fileServiceServer) GetFileStats(ctx context.Context, req *proto.FileRequest) (*proto.FileStats, error) {
	if len(req.FileData) == 0 {
		return &proto.FileStats{
			IsValid: false,
		}, nil
	}

	originalSize := int64(len(req.FileData))

	// Try to detect if it's a PDF
	isPDF := len(req.FileData) > 4 && string(req.FileData[:4]) == "%PDF"

	if isPDF {
		// Try to get page count
		pageCount, err := s.pdfProcessor.GetPageCount(req.FileData)
		if err != nil {
			log.Printf("Could not get page count: %v", err)
			return &proto.FileStats{
				OriginalSize: originalSize,
				PageCount:    0,
				FileType:     "pdf",
				IsValid:      false,
			}, nil
		}

		return &proto.FileStats{
			OriginalSize: originalSize,
			PageCount:    int32(pageCount),
			FileType:     "pdf",
			IsValid:      true,
		}, nil
	}

	return &proto.FileStats{
		OriginalSize: originalSize,
		FileType:     "unknown",
		IsValid:      false,
	}, nil
}

func main() {
	// Load .env file (ignore if not found)
	_ = godotenv.Load()

	// Create temp directory
	tempDir := os.TempDir()

	// Get port from environment or use default
	port := getPort()

	// Listen on port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", port, err)
	}

	// Create gRPC server with large message size limits
	maxMsgSize := 100 * 1024 * 1024 // 100MB
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	)
	fileServer := NewFileServiceServer(tempDir)

	// Register service
	proto.RegisterFileServiceServer(grpcServer, fileServer)

	// Handle graceful shutdown
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
		<-sigch
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
		os.Exit(0)
	}()

	// Start server
	fmt.Printf("File Service gRPC server listening on %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
