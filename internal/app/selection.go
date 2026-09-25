package app

import "fightogm/internal/ui"

func gridSelection(index, dx, dy, count int) int {
	if count <= 1 {
		return 0
	}
	cols := min(ui.SelectionColumns, count)
	rows := (count + cols - 1) / cols
	col := (index%cols + dx + cols) % cols
	row := (index/cols + dy + rows) % rows
	return min(count-1, row*cols+col)
}
