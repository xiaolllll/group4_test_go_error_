package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileProcessor 负责处理文件列表和文件读取
type FileProcessor struct {
	baseDir string
}

// NewFileProcessor 创建新的文件处理器
func NewFileProcessor(baseDir string) *FileProcessor {
	return &FileProcessor{
		baseDir: baseDir,
	}
}

// ProcessFileList 处理文件列表，搜索错误信息并写入输出文件
func (fp *FileProcessor) ProcessFileList(fileListPath, outputPath string, searcher *ErrorSearcher) (int, error) {
	// 读取文件列表
	filePaths, err := fp.readFileList(fileListPath)
	if err != nil {
		return 0, fmt.Errorf("读取文件列表失败: %w", err)
	}

	var allErrors []ErrorInfo
	errorCount := 0

	// 遍历文件列表中的每个文件
	for _, relativeFilePath := range filePaths {
		// 处理相对路径，确保去掉开头的 ./
		filePath := strings.TrimPrefix(relativeFilePath, "./")

		// 构建完整路径
		fullFilePath := filepath.Join(fp.baseDir, filePath)

		// 检查文件是否存在
		if _, err := os.Stat(fullFilePath); os.IsNotExist(err) {
			fmt.Printf("文件不存在: %s\n", fullFilePath)
			continue
		}

		// 只处理 .go 文件
		if !strings.HasSuffix(strings.ToLower(filePath), ".go") {
			continue
		}

		// 读取文件内容
		content, err := fp.readFileContent(fullFilePath)
		if err != nil {
			fmt.Printf("读取文件时出错 %s: %v\n", relativeFilePath, err)
			continue
		}

		// 搜索错误信息
		errors := searcher.SearchErrors(content, relativeFilePath)

		// 为每个错误分配索引
		for i := range errors {
			errorCount++
			errors[i].Index = errorCount
		}

		allErrors = append(allErrors, errors...)
	}

	// 写入输出文件
	if err := fp.writeOutputFile(outputPath, allErrors); err != nil {
		return 0, fmt.Errorf("写入输出文件失败: %w", err)
	}

	return errorCount, nil
}

// readFileList 读取文件列表
func (fp *FileProcessor) readFileList(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var filePaths []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			filePaths = append(filePaths, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return filePaths, nil
}

// readFileContent 读取文件内容
func (fp *FileProcessor) readFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// writeOutputFile 写入 Markdown 格式的输出文件
func (fp *FileProcessor) writeOutputFile(outputPath string, errors []ErrorInfo) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	header := "# 相关错误信息汇总\n\n"
	header += "| 报错日志 | 文件路径 | 行号 |\n"
	header += "| -------- | -------- | ---- |\n"

	if _, err := writer.WriteString(header); err != nil {
		return err
	}

	// 写入数据行
	for _, errInfo := range errors {
		// 对错误消息中的管道符进行处理，避免破坏表格格式
		escapedErrorMessage := strings.ReplaceAll(errInfo.ErrorMessage, "|", "\\|")
		line := fmt.Sprintf("| %s | %s | %d |\n", escapedErrorMessage, errInfo.FilePath, errInfo.LineNum)

		if _, err := writer.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

