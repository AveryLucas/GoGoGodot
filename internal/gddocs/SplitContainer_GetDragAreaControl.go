/*
$BarnacleButton.reparent($SplitContainer.get_drag_area_control())
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Button"
	"github.com/AveryLucas/gogogd/classdb/SplitContainer"
)

var barnacleButton Button.Instance
var splitContainer SplitContainer.Instance

func SplitContainer_GetDragAreaControl() {
	barnacleButton.AsNode().Reparent(splitContainer.GetDragAreaControl().AsNode())
}
