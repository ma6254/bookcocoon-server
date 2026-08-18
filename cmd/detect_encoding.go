package cmd

import (
	"fmt"
	"os"

	"github.com/saintfish/chardet"
	"github.com/spf13/cobra"
)

var detect_encoding_cmd = &cobra.Command{
	Use:   "detect-encoding [file]",
	Short: "检测文件的编码",
	Long:  `检测指定文件的文本编码`,
	Args:  cobra.MaximumNArgs(1), // 参数验证
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Hello\n")

		if len(args) < 1 {
			return fmt.Errorf("Please provide a file path")
		}

		fmt.Printf("input: %s\n", args[0])

		b, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}

		detector := chardet.NewTextDetector()

		result, err := detector.DetectAll(b)
		if err != nil {
			return err
		}

		for _, r := range result {
			fmt.Printf("Charset: %s, Language: %s, Confidence: %d\n", r.Charset, r.Language, r.Confidence)
		}

		return nil
	},
}

func init() {

	rootCmd.AddCommand(detect_encoding_cmd)
}
