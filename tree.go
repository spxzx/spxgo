package spxgo

import "strings"

// 前缀树
type treeNode struct {
	name       string
	routerName string
	isLeaf     bool
	children   []*treeNode
}

// Put path: /get/:id ... - > " " "get" ":id"
func (t *treeNode) Put(path string) {
	temp := t
	strs := strings.Split(path, "/")
	for index, name := range strs {
		if index == 0 { // 分割后 strs[0] = ' '
			continue
		}
		children := temp.children
		isMatch := false
		for _, node := range children {
			if node.name == name {
				isMatch = true
				temp = node
				break // 匹配路径成功跳出去查找下一个路径名
			}
		}
		// 没有匹配成功就在孩子节点这一层创建新的结点
		if !isMatch {
			isLeaf := false
			if index == len(strs)-1 {
				isLeaf = true
			}
			node := &treeNode{
				name:     name,
				children: make([]*treeNode, 0),
				isLeaf:   isLeaf,
			}
			children = append(children, node)
			temp.children = children
			temp = node
		}
	}
}

// Get path: /any/*/get ...
func (t *treeNode) Get(path string) *treeNode {
	temp, strs, routerName := t, strings.Split(path, "/"), ""
	for index, name := range strs {
		if index == 0 {
			continue
		}
		children := temp.children
		isMatch := false
		for _, node := range children {
			// 疑问点
			// 路由1 /test/:id
			// 路由2 /test/*/any
			// : 和 * 同时存在时该如何修改代码使得 : 不会取代 *
			if node.name == name ||
				node.name == "*" || // 有*号 在这里暂时理解为可以跳过这个结点去寻找下一个结点
				strings.Contains(node.name, ":") {
				isMatch = true
				routerName += "/" + node.name
				node.routerName = routerName
				temp = node
				// 到达最后一个结点，就将该结点返回
				if index == len(strs)-1 {
					return node
				}
				break
			}
		}
		if !isMatch {
			// 没有匹配就看有没有 ** 规则
			for _, node := range children {
				// /usr/** -> 所有路径
				// -> /usr/get
				// -> /usr/get/info
				if node.name == "**" {
					routerName += "/" + node.name
					node.routerName = routerName
					return node
				}
			}
		}
	}
	return nil
}
