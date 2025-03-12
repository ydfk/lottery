package middleware

import (
	"encoding/json"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v2"
)

// CamelCaseResponse 将所有响应转为小驼峰格式
func CamelCaseResponse() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 处理执行后的响应
		err := c.Next()

		// 获取响应体
		body := c.Response().Body()
		if len(body) > 0 {
			// 尝试将响应体解析为JSON
			var data interface{}
			if err := json.Unmarshal(body, &data); err == nil {
				// 转换为小驼峰格式
				camelCaseData := processObject(data)

				// 将转换后的数据重新序列化为JSON
				if camelCaseBody, err := json.Marshal(camelCaseData); err == nil {
					// 替换响应体
					c.Response().SetBody(camelCaseBody)
				}
			}
		}

		return err
	}
}

// convertToCamelCase 将结构体字段转换为小驼峰格式
func convertToCamelCase(data interface{}) interface{} {
	// 将数据转换为 JSON 字节
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return data // 如果无法转换，返回原始数据
	}

	// 转换为 map 或 map 数组
	var temp interface{}
	if err := json.Unmarshal(jsonBytes, &temp); err != nil {
		return data // 如果无法转换，返回原始数据
	}

	// 递归转换字段名为小驼峰
	result := processObject(temp)
	return result
}

// processObject 递归处理对象，将字段名转换为小驼峰
func processObject(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// 处理对象
		newMap := make(map[string]interface{})
		for key, value := range v {
			// 递归处理值
			processedValue := processObject(value)
			// 转换键为小驼峰并存储
			newMap[toCamelCase(key)] = processedValue
		}
		return newMap

	case []interface{}:
		// 处理数组
		newArray := make([]interface{}, len(v))
		for i, value := range v {
			// 递归处理数组中每个元素
			newArray[i] = processObject(value)
		}
		return newArray

	default:
		// 对于基本类型，直接返回
		return v
	}
}

// capitalize 首字母大写函数
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// toCamelCase 将字符串转换为小驼峰格式，如 "UserID" -> "userId"
func toCamelCase(s string) string {
	// 处理特殊情况
	if s == "" {
		return s
	}

	// 处理常见的特殊情况
	if s == "ID" {
		return "id"
	}

	// 简单处理：将首字母小写，保持其他字符不变
	if len(s) == 1 {
		return strings.ToLower(s)
	}

	// 拆分驼峰式命名
	var result []string
	var current []rune

	// 遍历字符
	for i, r := range s {
		// 如果是大写字母并且不是第一个字符
		if i > 0 && unicode.IsUpper(r) {
			// 保存当前单词并开始新的单词
			result = append(result, string(current))
			current = []rune{r}
		} else {
			// 继续当前单词
			current = append(current, r)
		}
	}

	// 添加最后一个单词
	if len(current) > 0 {
		result = append(result, string(current))
	}

	// 处理结果
	if len(result) == 0 {
		return s
	}

	// 第一个单词转为小写
	result[0] = strings.ToLower(result[0])

	// 后面的单词首字母大写
	for i := 1; i < len(result); i++ {
		result[i] = capitalize(strings.ToLower(result[i]))
	}

	// 连接所有单词
	return strings.Join(result, "")
}
