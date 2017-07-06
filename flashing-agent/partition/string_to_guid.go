package partition

import "fmt"

// Copied straight from github.com/rekby/gpt
func StringToGuid(guid string) (res [16]byte, err error) {
	byteOrder := [...]int{3, 2, 1, 0, -1, 5, 4, -1, 7, 6, -1, 8, 9, -1, 10, 11, 12, 13, 14, 15}
	if len(guid) != 36 {
		err = fmt.Errorf("BAD guid string length.")
		return
	}
	guidByteNum := 0
	for i := 0; i < len(guid); i += 2 {
		if byteOrder[guidByteNum] == -1 {
			if guid[i] == '-' {
				i++
				guidByteNum++
				if i >= len(guid)+1 {
					err = fmt.Errorf("BAD guid format minus")
					return
				}
			} else {
				err = fmt.Errorf("BAD guid char in minus pos")
				return
			}
		}

		sub := guid[i : i+2]
		var bt byte
		for pos, ch := range sub {
			var shift uint
			if pos == 0 {
				shift = 4
			} else {
				shift = 0
			}
			switch ch {
			case '0':
				bt |= 0 << shift
			case '1':
				bt |= 1 << shift
			case '2':
				bt |= 2 << shift
			case '3':
				bt |= 3 << shift
			case '4':
				bt |= 4 << shift
			case '5':
				bt |= 5 << shift
			case '6':
				bt |= 6 << shift
			case '7':
				bt |= 7 << shift
			case '8':
				bt |= 8 << shift
			case '9':
				bt |= 9 << shift
			case 'A', 'a':
				bt |= 10 << shift
			case 'B', 'b':
				bt |= 11 << shift
			case 'C', 'c':
				bt |= 12 << shift
			case 'D', 'd':
				bt |= 13 << shift
			case 'E', 'e':
				bt |= 14 << shift
			case 'F', 'f':
				bt |= 15 << shift
			default:
				err = fmt.Errorf("BAD guid char: ", i+pos, ch)
				return
			}
		}
		res[byteOrder[guidByteNum]] = bt
		guidByteNum++
	}
	return res, nil
}
