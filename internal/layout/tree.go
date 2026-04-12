package layout

import "goglweb/internal/style"

func BuildLayoutTree(styledNode *style.StyledNode) *LayoutBox {
	root := buildBoxTree(styledNode)
	return root
}

func buildBoxTree(styledNode *style.StyledNode) *LayoutBox {
	var boxType BoxType
	display := styledNode.Display()

	switch display {
	case style.Block:
		boxType = BlockBox
	case style.Inline:
		boxType = InlineBox
	case style.None:
		return nil
	default:
		boxType = InlineBox
	}

	box := &LayoutBox{
		BoxType:    boxType,
		StyledNode: styledNode,
		Dimensions: Dimensions{},
		Children:   []*LayoutBox{},
	}

	for _, child := range styledNode.Children {
		childBox := buildBoxTree(child)
		if childBox == nil {
			continue
		}

		if needsAnonymousBox(box, childBox) {
			var anonymousBox *LayoutBox
			if len(box.Children) > 0 && box.Children[len(box.Children)-1].BoxType == AnonymousBox {
				anonymousBox = box.Children[len(box.Children)-1]
			} else {
				anonymousBox = &LayoutBox{
					BoxType:    AnonymousBox,
					StyledNode: nil,
					Dimensions: Dimensions{},
					Children:   []*LayoutBox{},
					Parent:     box,
				}
				box.Children = append(box.Children, anonymousBox)
			}
			childBox.Parent = anonymousBox
			anonymousBox.Children = append(anonymousBox.Children, childBox)
		} else {
			childBox.Parent = box
			box.Children = append(box.Children, childBox)
		}
	}

	return box
}

func needsAnonymousBox(parent *LayoutBox, child *LayoutBox) bool {
	return parent.BoxType == BlockBox && child.BoxType == InlineBox
}
