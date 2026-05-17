package testkit

import (
	"fmt"
	"net/textproto"
	"strings"
)

type textprotoMIMEHeader = textproto.MIMEHeader

func multipartFileContentDisposition(fieldName, fileName string) string {
	return fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldName), escapeQuotes(fileName))
}

func escapeQuotes(s string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(s)
}
