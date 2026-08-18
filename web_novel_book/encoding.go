package web_novel_book

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/saintfish/chardet"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// 建立 chardet 返回名称 与 x/text 编码对象的映射
var charsetMap = map[string]encoding.Encoding{
	"UTF-8":        unicode.UTF8,
	"UTF-16BE":     unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM),
	"UTF-16LE":     unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
	"GBK":          simplifiedchinese.GBK,
	"GB-18030":     simplifiedchinese.GB18030,
	"Big-5":        traditionalchinese.Big5,
	"Shift_JIS":    japanese.ShiftJIS,
	"EUC-JP":       japanese.EUCJP,
	"EUC-KR":       korean.EUCKR,
	"ISO-8859-1":   charmap.ISO8859_1,
	"windows-1252": charmap.Windows1252,
}

type TitleRegexp struct {
	Name  string
	Regex string
}

var TitleRegexpList = []TitleRegexp{
	{

		// 第零章 【伯爵的儿子】
		// 第一章 【白痴】
		// 第二章【文不成，武不就】（上）
		// 第二章 【文不成，武不就】（下）
		// 第三章 【魔法的道路】

		Name:  "第(汉字数字)章",
		Regex: `(?m)^第[零一二三四五六七八九十百千万]+章.*$`,
	},
	{
		Name:  "第(数字)章",
		Regex: `(?m)^[ ]*第[0-9]+章.*$`,
	},
}

// 将任意编码的字节数据转换为 UTF-8
func ConvertCharsetToUTF8(data []byte, charset string) ([]byte, error) {
	// 大小写统一，chardet 有时返回 "utf-8" 小写
	charset = strings.ToUpper(charset)

	// 如果已经是 UTF-8，直接返回（chardet 对纯英文可能误判为 ASCII，这里也兼容）
	if charset == "UTF-8" || charset == "ASCII" {
		return data, nil
	}

	enc, ok := charsetMap[charset]
	if !ok {
		return nil, fmt.Errorf("不支持的编码格式: %s", charset)
	}

	// 创建解码器并进行转换
	decoder := enc.NewDecoder()
	utf8Data, err := io.ReadAll(transform.NewReader(bytes.NewReader(data), decoder))
	if err != nil {
		return nil, fmt.Errorf("转换失败: %w", err)
	}
	return utf8Data, nil
}

// 将任意编码的字节数据转换为 UTF-8
func ConvertAnyToUTF8(input []byte) ([]byte, error) {

	detector := chardet.NewTextDetector()
	results, err := detector.DetectAll(input)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("无法检测文件编码")
	}

	result := results[0]

	if result.Confidence != 100 {
		fmt.Printf("Warning: The detected encoding may not be accurate. Detected encoding: %s, Confidence: %d\n", result.Charset, result.Confidence)
	}

	charset := result.Charset

	// 统一大写
	charset = strings.ToUpper(charset)

	// 如果已经是 UTF-8，直接返回（chardet 对纯英文可能误判为 ASCII，这里也兼容）
	if charset == "UTF-8" {
		return input, nil
	}

	enc, ok := charsetMap[charset]
	if !ok {
		return nil, fmt.Errorf("不支持的编码格式: %s", charset)
	}

	// 创建解码器并进行转换
	decoder := enc.NewDecoder()
	utf8Data, err := io.ReadAll(transform.NewReader(bytes.NewReader(input), decoder))
	if err != nil {
		return nil, fmt.Errorf("转换失败: %w", err)
	}
	return utf8Data, nil
}

type Chapter struct {
	Title   string
	Content string
}

// 将任意编码的字节数据转换为 UTF-8
func SplitChapter(input []byte) ([]*Chapter, error) {

	isMatched := false
	results := [][]int{}

	for _, titleRegexp := range TitleRegexpList {

		re := regexp.MustCompile(titleRegexp.Regex)
		if re == nil {
			return nil, fmt.Errorf("无法编译正则表达式: %s", titleRegexp.Regex)
		}

		results = re.FindAllSubmatchIndex(input, -1)

		if len(results) > 0 {
			isMatched = true
			break
		}
	}

	if !isMatched {
		return nil, nil
	}

	chapters := []*Chapter{}

	for i := 0; i < len(results); i++ {

		match_start := results[i][0]
		match_end := results[i][1]

		title := string(input[match_start:match_end])
		content := ""

		if i == len(results)-1 {
			// 最后一章，内容到文件末尾
			content = string(input[match_end:])
		} else {
			content = string(input[match_end:results[i+1][0]])
		}

		content = strings.TrimSpace(content)
		if len(content) == 0 {
			continue
		}

		// log.Printf("%s", title)

		chapter := &Chapter{
			Title:   title,
			Content: content,
		}

		chapters = append(chapters, chapter)
	}

	return chapters, nil
}
