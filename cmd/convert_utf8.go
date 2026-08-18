package cmd

import (
	"fmt"
	"os"

	"github.com/ma6254/bookcocoon-server/web_novel_book"
	"github.com/saintfish/chardet"
	"github.com/spf13/cobra"
)

var convert_utf8 = &cobra.Command{
	Use:   "convert-utf8 [file]",
	Short: "将文件转换为 UTF-8 编码",
	Long:  `将指定文件的文本编码转换为 UTF-8`,
	Args:  cobra.MaximumNArgs(2), // 参数验证
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Hello\n")

		if len(args) < 1 {
			return fmt.Errorf("Please provide a file path")
		}

		input_file_path := args[0]
		output_file_path := args[1]

		fmt.Printf("input: %s\n", input_file_path)
		fmt.Printf("output: %s\n", output_file_path)

		b, err := os.ReadFile(input_file_path)
		if err != nil {
			return err
		}

		detector := chardet.NewTextDetector()

		result, err := detector.DetectAll(b)
		if err != nil {
			return err
		}

		if len(result) == 0 {
			return fmt.Errorf("无法检测文件编码")
		}

		for _, r := range result {
			fmt.Printf("Charset: %s, Language: %s, Confidence: %d\n", r.Charset, r.Language, r.Confidence)
		}

		// 如果检测结果的置信度不为 100%，则提示用户可能存在误判
		if result[0].Confidence != 100 {
			fmt.Printf("Warning: The detected encoding may not be accurate. Detected encoding: %s, Confidence: %d\n", result[0].Charset, result[0].Confidence)
			return nil
		}

		// 将文件内容转换为 UTF-8
		utf8Data, err := web_novel_book.ConvertCharsetToUTF8(b, result[0].Charset)
		if err != nil {
			return err
		}

		// 将转换后的内容写回文件
		err = os.WriteFile(output_file_path, utf8Data, 0644)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {

	rootCmd.AddCommand(convert_utf8)
}
