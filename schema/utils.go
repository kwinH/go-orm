package schema

import "strings"

func SnakeString(s string) string {
	var builder strings.Builder
	builder.Grow(len(s) * 2)

	for i, d := range s {
		if i > 0 && d >= 'A' && d <= 'Z' {
			builder.WriteByte('_')
		}
		builder.WriteRune(d)
	}

	return strings.ToLower(builder.String())
}

func ParseTagSetting(str string, sep string) map[string]string {
	settings := make(map[string]string)
	names := strings.Split(str, sep)
	var buffer strings.Builder

	for _, name := range names {
		// 处理以 '\' 结尾的转义符
		if len(name) > 0 && name[len(name)-1] == '\\' {
			if buffer.Len() > 0 {
				buffer.WriteString(sep)
			}
			buffer.WriteString(name[:len(name)-1])
			continue
		}

		// 将缓冲区中的内容与当前name组合
		if buffer.Len() > 0 {
			buffer.WriteString(sep)
			buffer.WriteString(name)
			name = buffer.String()
			buffer.Reset()
		}

		values := strings.SplitN(name, ":", 2)
		key := strings.TrimSpace(values[0])

		if len(values) == 2 {
			settings[key] = values[1]
		} else if key != "" {
			settings[key] = ""
		}
	}

	return settings
}
