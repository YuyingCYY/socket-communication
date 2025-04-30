package models

import (
	"os"
)

type FileResponseHandler func(
	jsonData []byte,
	dataData []byte,
	jsonFileInfo os.FileInfo,
	dataFileInfo os.FileInfo,
	req Request,
) Response
