package trie

const (
	WildCard = '*'
)

// 字典树节点
type TrieNode struct {
	children map[byte]*TrieNode
	isEnd    bool
}

// 构造字典树节点
func newTrieNode() *TrieNode {
	return &TrieNode{children: make(map[byte]*TrieNode), isEnd: false}
}

// 字典树
type Trie struct {
	root *TrieNode
}

// 构造字典树
func NewTrie() *Trie {
	return &Trie{root: newTrieNode()}
}

// 向字典树中插入一个单词，大小写敏感
func (trie *Trie) Insert(word []byte) {
	node := trie.root
	for i := 0; i < len(word); i++ {
		_, ok := node.children[word[i]]
		if !ok {
			node.children[word[i]] = newTrieNode()
		}
		node = node.children[word[i]]
	}
	node.isEnd = true
}

// 搜索字典树中是否存在指定单词，匹配前缀即可
func (trie *Trie) Search(word []byte) bool {
	node := trie.root
	for i := 0; i < len(word); i++ {
		_, ok := node.children[word[i]]
		if !ok {
			if node.isEnd {
				return true
			}
			return false
		}
		node = node.children[word[i]]
	}
	return node.isEnd
}
