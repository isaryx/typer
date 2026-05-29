package defense

// ShieldRow is the row index where falling words are destroyed.
func ShieldRow() int {
	return PlayfieldRows - BaseArtRows - 1
}

// BaseArtStartRow is the first playfield row of the base art block.
func BaseArtStartRow() int {
	return ShieldRow() + 1
}
