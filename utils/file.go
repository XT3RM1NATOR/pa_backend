package utils

import "errors"

func extractFileType(binaryData []byte) (string, error) {
	if len(binaryData) < 4 {
		return "", errors.New("unknown data type")
	}

	// Check the magic number to identify file type
	if binaryData[0] == 0x47 && binaryData[1] == 0x49 && binaryData[2] == 0x46 && binaryData[3] == 0x38 {
		return ".gif", nil
	} else if binaryData[0] == 0x49 && binaryData[1] == 0x44 && binaryData[2] == 0x33 { // "ID3" in ASCII
		return ".mp3", nil
	} else if binaryData[0] == 0x66 && binaryData[1] == 0x74 && binaryData[2] == 0x79 && binaryData[3] == 0x70 { // "ftyp" in ASCII
		return ".mp4", nil
	} else if binaryData[0] == 0xFF && binaryData[1] == 0xD8 && binaryData[2] == 0xFF {
		return ".jpeg", nil
	} else if binaryData[0] == 0x00 && binaryData[1] == 0x00 && binaryData[2] == 0x00 && binaryData[3] == 0x18 && binaryData[16] == 0x66 && binaryData[17] == 0x74 && binaryData[18] == 0x79 && binaryData[19] == 0x70 { // "ftyp" in ASCII
		return ".heic", nil
	} else if binaryData[0] == 0x42 && binaryData[1] == 0x4D {
		return ".bmp", nil
	} else if binaryData[0] == 0x49 && binaryData[1] == 0x49 && binaryData[2] == 0x2A && binaryData[3] == 0x00 {
		return ".tiff", nil
	} else if binaryData[0] == 0x4D && binaryData[1] == 0x4D && binaryData[2] == 0x00 && binaryData[3] == 0x2A {
		return ".tiff", nil
	} else if binaryData[0] == 0x54 && binaryData[1] == 0x45 && binaryData[2] == 0x58 && binaryData[3] == 0x54 {
		return ".txt", nil
	}

	return "", nil
}
