package serializer

// checkedTableLayout validates an untrusted fixed-record table before callers
// convert its uint32 header fields to int, allocate result slices, or index the
// payload. The division form avoids count*recordSize integer overflow.
func checkedTableLayout(count, recordSize uint32, dataLen, minRecordSize int) (int, int, bool) {
	if count == 0 || recordSize == 0 || dataLen < 0 || minRecordSize < 0 {
		return 0, 0, false
	}
	if uint64(recordSize) < uint64(minRecordSize) {
		return 0, 0, false
	}
	if uint64(count) > uint64(dataLen)/uint64(recordSize) {
		return 0, 0, false
	}
	return int(count), int(recordSize), true
}
