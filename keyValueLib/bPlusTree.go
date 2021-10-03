package main

import (
	"fmt"
	"sort"
)

type bPlusTree struct {
	maxChildren      int
	maxKeys          int
	rootInternalNode *internalNode
	firstLeafNode    *leafNode
	lastLeafNode     *leafNode
	internalNodes    []*internalNode
	leafNodes        []*leafNode
}

type internalNode struct {
	keys               []internalNodeKey
	parentInternalNode *internalNode
	lastKey            internalNodeKey
}

type internalNodeKey struct {
	key         string
	treePointer *internalNode
	leafPointer *leafNode
}

type leafNode struct {
	keys []leafNodeKey
	prev *leafNode
	next *leafNode
}

type leafNodeKey struct {
	key  string
	data string
}

func setup() *bPlusTree {
	newTree := &bPlusTree{
		maxChildren: 501,
		maxKeys:     3,
	}

	newLeafNode := &leafNode{}

	newTree.leafNodes = append(newTree.leafNodes, newLeafNode)
	newTree.firstLeafNode = newLeafNode
	newTree.lastLeafNode = newLeafNode

	return newTree
}

func (bpt *bPlusTree) add(key leafNodeKey) {
	if len(bpt.internalNodes) == 0 {
		bpt.firstLeafNode.keys = append(bpt.firstLeafNode.keys, key)
		sort.Slice(bpt.firstLeafNode.keys, func(i, j int) bool {
			return bpt.firstLeafNode.keys[i].key < bpt.firstLeafNode.keys[j].key
		})

		keysLen := len(bpt.firstLeafNode.keys)
		if keysLen >= bpt.maxKeys {
			median := int(keysLen / 2)
			medianKey := bpt.firstLeafNode.keys[median].key

			newInternalNode := &internalNode{}
			newInternalNode.keys = append(newInternalNode.keys,
				internalNodeKey{key: medianKey, leafPointer: bpt.firstLeafNode})

			bpt.internalNodes = append(bpt.internalNodes, newInternalNode)
			bpt.rootInternalNode = newInternalNode

			newLeafNode := &leafNode{prev: bpt.firstLeafNode}
			newLeafNode.keys = append(newLeafNode.keys, bpt.firstLeafNode.keys[median:]...)

			bpt.leafNodes = append(bpt.leafNodes, newLeafNode)
			bpt.lastLeafNode = newLeafNode
			newInternalNode.lastKey = internalNodeKey{leafPointer: newLeafNode}

			bpt.firstLeafNode.keys = bpt.firstLeafNode.keys[:median]
			bpt.firstLeafNode.next = newLeafNode
		}
	} else {
		bpt.insertKey(key)
	}
}

func (bpt *bPlusTree) insertKey(lnk leafNodeKey) {

	currentNode := bpt.rootInternalNode

	found := false

	for {

		found = false

		for _, key := range currentNode.keys {
			if lnk.key < key.key {
				found = true
				if key.leafPointer != nil {
					bpt.insertIntoLeafNode(lnk, key.leafPointer, currentNode)
					return
				} else if key.treePointer != nil {
					currentNode = key.treePointer
					break
				}
			}
		}

		if !found {
			if currentNode.lastKey.leafPointer != nil {
				bpt.insertIntoLeafNode(lnk, currentNode.lastKey.leafPointer, currentNode)
				return
			} else if currentNode.lastKey.treePointer != nil {
				currentNode = currentNode.lastKey.treePointer
			}
		}

	}

}

func (bpt *bPlusTree) insertIntoLeafNode(lnk leafNodeKey, ln *leafNode, currentNode *internalNode) {
	ln.keys = append(ln.keys, lnk)
	sort.Slice(ln.keys, func(i, j int) bool {
		return ln.keys[i].key < ln.keys[j].key
	})

	keysLen := len(ln.keys)

	if keysLen >= bpt.maxKeys {
		median := int(keysLen / 2)
		medianKey := ln.keys[median].key

		newLeafNode := &leafNode{prev: ln, next: ln.next}
		newLeafNode.keys = append(newLeafNode.keys, ln.keys[median:]...)
		bpt.leafNodes = append(bpt.leafNodes, newLeafNode)

		if ln.next == nil {
			bpt.lastLeafNode = newLeafNode
		}

		ln.next = newLeafNode
		ln.keys = ln.keys[:median]

		keysLen := len(currentNode.keys)
		if keysLen > 0 && currentNode.keys[keysLen-1].key > medianKey { // swap them round if not adding to end
			*ln, *newLeafNode = *newLeafNode, *ln
		}

		bpt.insertIntoInternalNode(medianKey, newLeafNode, nil, currentNode)
	}
}

func (bpt *bPlusTree) insertIntoInternalNode(medianKey string, leafPointer *leafNode, treePointer *internalNode, currentNode *internalNode) *internalNode {
	newInternalNodeKey := internalNodeKey{
		key:         medianKey,
		leafPointer: leafPointer,
		treePointer: treePointer,
	}

	currentNodeKeysLen := len(currentNode.keys)

	if currentNodeKeysLen > 0 && currentNode.keys[currentNodeKeysLen-1].key < medianKey {
		savedKey := currentNode.lastKey
		savedKey.key = medianKey
		currentNode.keys = append(currentNode.keys, savedKey)
		newInternalNodeKey.key = ""
		currentNode.lastKey = newInternalNodeKey
	} else {
		currentNode.keys = append(currentNode.keys, newInternalNodeKey)
	}

	sort.Slice(currentNode.keys, func(i, j int) bool {
		return currentNode.keys[i].key < currentNode.keys[j].key
	})

	inLen := len(currentNode.keys)
	if inLen >= bpt.maxKeys {
		median := int(inLen / 2)
		newMedianKey := currentNode.keys[median].key

		newInternalNode := &internalNode{
			lastKey: currentNode.lastKey,
		}
		newInternalNode.keys = append(newInternalNode.keys, currentNode.keys[median+1:]...)
		bpt.internalNodes = append(bpt.internalNodes, newInternalNode)

		currentNode.lastKey = currentNode.keys[median]
		currentNode.lastKey.key = ""
		currentNode.keys = currentNode.keys[:median]

		if currentNode.parentInternalNode == nil {
			newRootInternalNode := &internalNode{lastKey: internalNodeKey{
				treePointer: newInternalNode,
			}}
			bpt.internalNodes = append(bpt.internalNodes, newRootInternalNode)

			bpt.rootInternalNode = newRootInternalNode
			currentNode.parentInternalNode = newRootInternalNode
			newInternalNode.parentInternalNode = newRootInternalNode

			bpt.insertIntoInternalNode(newMedianKey, nil, currentNode, currentNode.parentInternalNode)
		} else {
			parent := currentNode.parentInternalNode
			parentKeysLen := len(parent.keys)
			if parentKeysLen > 0 && parent.keys[parentKeysLen-1].key > medianKey { // swap them round if not adding to end
				*currentNode, *newInternalNode = *newInternalNode, *currentNode
			}
			newParent := bpt.insertIntoInternalNode(newMedianKey, nil, newInternalNode, parent)
			currentNode.parentInternalNode = newParent
			newInternalNode.parentInternalNode = newParent
		}

		return newInternalNode
	}

	return currentNode
}

func (tree *bPlusTree) printTest() {

	queue := tree.rootInternalNode.keys[:]
	queue = append(queue, tree.rootInternalNode.lastKey)

	fmt.Println("root", tree.rootInternalNode.keys, tree.rootInternalNode.lastKey)
	for {
		var key internalNodeKey
		key, queue = queue[0], queue[1:]

		if len(queue) == 0 {
			fmt.Println("leaf", tree.lastLeafNode.keys)
			break
		}

		if key.treePointer != nil {
			fmt.Println("under", key.key, key.treePointer.keys, key.treePointer.lastKey)

			queue = append(queue, key.treePointer.keys...)
			queue = append(queue, key.treePointer.lastKey)
		}

		if key.leafPointer != nil {

			fmt.Println("(leaf) under", key.key, key.leafPointer.keys)

		}

	}
}

func (bpt *bPlusTree) find(keyToFind string) string {

	currentNode := bpt.rootInternalNode

	found := false

	for {

		found = false

		for _, key := range currentNode.keys {
			if keyToFind < key.key {
				found = true
				if key.leafPointer != nil {
					return searchForKey(keyToFind, key.leafPointer.keys)
				} else if key.treePointer != nil {
					currentNode = key.treePointer
					break
				}
			}
		}

		if !found {
			if currentNode.lastKey.leafPointer != nil {
				return searchForKey(keyToFind, currentNode.lastKey.leafPointer.keys)
			} else if currentNode.lastKey.treePointer != nil {
				currentNode = currentNode.lastKey.treePointer
			}
		}
	}

}

func searchForKey(str string, data []leafNodeKey) string {
	index := sort.Search(len(data), func(i int) bool { return data[i].key >= str })
	return data[index].data
}

func (bpt *bPlusTree) findPath(keyToFind string) {

	type loopStruct struct {
		keys []internalNodeKey
		node *internalNode
	}

	var queue []loopStruct

	queue = append(queue, loopStruct{
		keys: bpt.rootInternalNode.keys,
		node: bpt.rootInternalNode,
	})

	for {
		var item loopStruct
		item, queue = queue[0], queue[1:]

		found := false

		for _, key := range item.keys {

			if keyToFind < key.key {
				found = true
				if key.leafPointer != nil {
					fmt.Print(key.key)
					break
				} else if key.treePointer != nil {
					fmt.Print(key.key)
					queue = append(queue, loopStruct{
						keys: key.treePointer.keys,
						node: key.treePointer,
					})
					break
				}
			}
		}

		if found {
			break
		}

		if item.node.lastKey.treePointer != nil {
			fmt.Print(item.node.keys[len(item.node.keys)-1].key)
			queue = append(queue, loopStruct{
				keys: item.node.lastKey.treePointer.keys,
				node: item.node.lastKey.treePointer,
			})
		}

		if len(queue) == 0 {
			fmt.Print(item.node.keys[len(item.node.keys)-1].key)
			break
		}

		fmt.Print(" -> ")
	}
	fmt.Println("")
}

func main() {
	tree := setup()

	tree.add(leafNodeKey{key: "h", data: "THIS IS DATA H"})
	tree.add(leafNodeKey{key: "i", data: "THIS IS DATA I"})
	tree.add(leafNodeKey{key: "j", data: "THIS IS DATA J"})
	tree.add(leafNodeKey{key: "k", data: "THIS IS DATA K"})
	tree.add(leafNodeKey{key: "l", data: "THIS IS DATA L"})
	tree.add(leafNodeKey{key: "m", data: "THIS IS DATA M"})
	tree.add(leafNodeKey{key: "a", data: "THIS IS DATA A"})
	tree.add(leafNodeKey{key: "b", data: "THIS IS DATA B"})
	tree.add(leafNodeKey{key: "c", data: "THIS IS DATA C"})
	tree.add(leafNodeKey{key: "d", data: "THIS IS DATA D"})
	tree.add(leafNodeKey{key: "e", data: "THIS IS DATA E"})
	tree.add(leafNodeKey{key: "f", data: "THIS IS DATA F"})
	tree.add(leafNodeKey{key: "g", data: "THIS IS DATA G"})

	tree.findPath("g")

	fmt.Println(tree.find("g"))

	tree.printTest()
}
