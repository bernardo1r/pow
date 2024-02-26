package save

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
)

type Entry struct {
	FileDigest []byte
	Challenge  []byte
	PowDigest  []byte
	Zeros      int
}

type EntrySet map[string]*Entry

type EntryHex struct {
	FileDigest string `json:"fileDigest"`
	Challenge  string `json:"challenge"`
	PowDigest  string `json:"powDigest"`
	Zeros      int    `json:"zeros"`
}

func (e *Entry) UnmarshalJSON(data []byte) error {
	entryHex := new(EntryHex)
	err := json.Unmarshal(data, &entryHex)
	if err != nil {
		return err
	}

	e.FileDigest, err = hex.DecodeString(entryHex.FileDigest)
	if err != nil {
		return err
	}
	e.Challenge, err = hex.DecodeString(entryHex.Challenge)
	if err != nil {
		return err
	}
	e.PowDigest, err = hex.DecodeString(entryHex.PowDigest)
	if err != nil {
		return err
	}
	e.Zeros = entryHex.Zeros

	return nil
}

func (e *Entry) MarshalJSON() ([]byte, error) {
	entryHex := new(EntryHex)
	entryHex.FileDigest = hex.EncodeToString(e.FileDigest)
	entryHex.Challenge = hex.EncodeToString(e.Challenge)
	entryHex.PowDigest = hex.EncodeToString(e.PowDigest)
	entryHex.Zeros = e.Zeros
	data, err := json.Marshal(entryHex)
	return data, err
}

func checkEmptyFile(file *os.File) (bool, error) {
	info, err := file.Stat()
	if err != nil {
		return false, err
	}

	if info.Size() > 0 {
		return false, nil
	}
	return true, nil
}

func SetFromJSONFile(file *os.File) (EntrySet, error) {
	set := make(EntrySet)
	ok, err := checkEmptyFile(file)
	if err != nil {
		return nil, err
	}
	if ok {
		return set, nil
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&set)
	return set, err
}

func (s *EntrySet) EncodeJSONFile(file *os.File) error {
	err := file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(s)
}

func Save(outputFilename string, targetFilename string, result *Entry) error {
	file, err := os.OpenFile(outputFilename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	set, err := SetFromJSONFile(file)
	if err != nil {
		return err
	}
	set[targetFilename] = result

	return set.EncodeJSONFile(file)
}
