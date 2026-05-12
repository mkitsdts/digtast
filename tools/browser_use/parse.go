package browser_use

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func PreCleanHTML(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}

	// 1. 解析片段
	context := &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	}
	nodes, err := html.ParseFragment(strings.NewReader(input), context)
	if err != nil {
		return input
	}

	var buf bytes.Buffer
	for _, n := range nodes {
		if cleaned := processNode(n); cleaned != nil {
			// 只有非空节点才渲染
			html.Render(&buf, cleaned)
		}
	}

	// 最后对整体结果做一次 Trim，彻底去掉由于标签删除留下的换行符
	return strings.TrimSpace(buf.String())
}

func processNode(n *html.Node) *html.Node {
	// 1. 基础过滤：移除注释和空白文本节点
	if n.Type == html.CommentNode {
		return nil
	}
	if n.Type == html.TextNode {
		if strings.TrimSpace(n.Data) == "" {
			return nil
		}
		return n
	}

	if n.Type == html.ElementNode {
		tagName := strings.ToLower(n.Data)

		// 2. 强力黑名单：这些标签及其子标签全部不要
		noiseTags := map[string]bool{
			"script": true, "style": true, "meta": true, "link": true,
			"head": true, "title": true, "svg": true, "canvas": true,
			"noscript": true, "iframe": true,
		}
		if noiseTags[tagName] {
			return nil
		}

		// 3. 属性过滤：只保留可能具有交互或定位意义的属性
		keepAttrs := []html.Attribute{}
		for _, attr := range n.Attr {
			name := strings.ToLower(attr.Key)
			// 保留 id, name, href, type, placeholder, value, role
			if name == "id" || name == "name" || name == "href" ||
				name == "type" || name == "placeholder" || name == "role" || name == "value" {
				keepAttrs = append(keepAttrs, attr)
			}
		}
		n.Attr = keepAttrs
	}

	// 4. 递归处理子节点
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		if cleanedChild := processNode(c); cleanedChild == nil {
			n.RemoveChild(c)
		}
		c = next
	}

	// 5. 关键：剪枝逻辑
	// 如果是容器标签（div, span, p 等），清理后如果没有子节点且没有文本，则判定为不可操作，直接删除
	if n.Type == html.ElementNode {
		if n.FirstChild == nil {
			// 如果是交互性标签（如 input, img, a），即使没子节点也保留
			interactiveTags := map[string]bool{
				"input": true, "button": true, "a": true, "img": true, "textarea": true, "select": true,
			}
			if !interactiveTags[strings.ToLower(n.Data)] {
				return nil
			}
		}
	}

	return n
}
