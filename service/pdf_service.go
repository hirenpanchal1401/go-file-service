package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type PDFProcessor struct {
	TempDir string
}

func NewPDFProcessor(tempDir string) *PDFProcessor {
	return &PDFProcessor{
		TempDir: tempDir,
	}
}

// RepairPDF attempts to repair corrupted PDFs and compress them
func (p *PDFProcessor) RepairPDF(fileData []byte) ([]byte, error) {
	inputFile, err := p.saveToTempFile(fileData, "input_*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to save input file: %w", err)
	}
	defer os.Remove(inputFile)

	outputFile := inputFile + "_optimized.pdf"
	defer os.Remove(outputFile)

	// Validate and optimize the PDF - pdfcpu will attempt to repair during optimization
	if err := api.OptimizeFile(inputFile, outputFile, nil); err != nil {
		log.Printf("PDF optimization failed: %v, trying to read original", err)
		// Return original if optimization fails
		return fileData, nil
	}

	// Read the optimized PDF
	optimizedData, err := os.ReadFile(outputFile)
	if err != nil {
		log.Printf("Failed to read optimized file: %v", err)
		return fileData, nil
	}

	// If optimized version is larger, return original
	if len(optimizedData) > len(fileData) {
		return fileData, nil
	}

	return optimizedData, nil
}

// CompressPDF compresses a valid PDF
func (p *PDFProcessor) CompressPDF(fileData []byte) ([]byte, error) {
	inputFile, err := p.saveToTempFile(fileData, "input_*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to save input file: %w", err)
	}
	defer os.Remove(inputFile)

	outputFile := inputFile + "_compressed.pdf"
	defer os.Remove(outputFile)

	// Optimize (compress) the PDF
	if err := api.OptimizeFile(inputFile, outputFile, nil); err != nil {
		return nil, fmt.Errorf("failed to compress PDF: %w", err)
	}

	return os.ReadFile(outputFile)
}

// ValidatePDF checks if a PDF is valid
func (p *PDFProcessor) ValidatePDF(fileData []byte) error {
	inputFile, err := p.saveToTempFile(fileData, "validate_*.pdf")
	if err != nil {
		return err
	}
	defer os.Remove(inputFile)

	return api.ValidateFile(inputFile, nil)
}

// GetPageCount returns the number of pages in a PDF
func (p *PDFProcessor) GetPageCount(fileData []byte) (int, error) {
	inputFile, err := p.saveToTempFile(fileData, "pagecount_*.pdf")
	if err != nil {
		return 0, err
	}
	defer os.Remove(inputFile)

	ctx, err := api.ReadContextFile(inputFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read PDF: %w", err)
	}

	return ctx.PageCount, nil
}

// saveToTempFile saves bytes to a temporary file
func (p *PDFProcessor) saveToTempFile(data []byte, pattern string) (string, error) {
	f, err := os.CreateTemp(p.TempDir, pattern)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, bytes.NewReader(data)); err != nil {
		os.Remove(f.Name())
		return "", err
	}

	return f.Name(), nil
}

// CalculateCompressionRatio calculates the compression ratio
func CalculateCompressionRatio(originalSize, compressedSize int64) float32 {
	if originalSize == 0 {
		return 0
	}
	return float32((originalSize - compressedSize)) / float32(originalSize) * 100
}
