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
	case style.Flex:
		boxType = FlexBox
	case style.None:
		return nil
	default:
		boxType = InlineBox
	}

	box := &LayoutBox{
		BoxType:      boxType,
		StyledNode:   styledNode,
		Dimensions:   Dimensions{},
		Children:     []*LayoutBox{},
		PositionType: "static",
	}

	// Read position property and adjust BoxType/PositionType accordingly
	pos := ""
	if styledNode.SpecifiedValues != nil {
		pos = styledNode.SpecifiedValues["position"]
	}
	switch pos {
	case "relative":
		box.PositionType = "relative"
		// relative stays in normal flow (BlockBox), just shifted later
	case "absolute":
		box.BoxType = PositionedBox
		box.PositionType = "absolute"
	case "fixed":
		box.BoxType = PositionedBox
		box.PositionType = "fixed"
	}

	// Read overflow property
	if styledNode.SpecifiedValues != nil {
		if overflow, ok := styledNode.SpecifiedValues["overflow"]; ok {
			box.Overflow = overflow
		}
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
	if parent.BoxType == FlexBox {
		return false
	}
	return parent.BoxType == BlockBox && child.BoxType == InlineBox
}
