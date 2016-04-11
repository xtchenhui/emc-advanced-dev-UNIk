package compilers

import (
	"encoding/json"
	"io"

	uniktypes "github.com/emc-advanced-dev/unik/pkg/types"

	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// uses rump docker conter container
// the container expectes code in /opt/code and will produce program.bin in the same folder.
// we need to take the program bin and combine with json config produce an image

type RunmpCompiler struct {
	DockerImage string
	CreateImage func(kernel, args string, mntPoints []string) (*uniktypes.RawImage, error)
}

func (r *RunmpCompiler) CompileRawImage(sourceTar io.ReadCloser, args string, mntPoints []string) (*uniktypes.RawImage, error) {

	localFolder, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(localFolder)

	if err := ExtractTar(sourceTar, localFolder); err != nil {
		return nil, err
	}

	if err := r.runContainer(localFolder); err != nil {
		return nil, err
	}

	// now we should program.bin
	resultFile := path.Join(localFolder, "program.bin")

	return r.CreateImage(resultFile, args, mntPoints)
}

// rump special json
func ToRumpJson(c RumpConfig) (string, error) {

	blk := c.Blk
	c.Blk = nil

	jsonConfig, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	blks := ""
	for _, b := range blk {

		blkjson, err := json.Marshal(b)
		if err != nil {
			return "", err
		}
		blks += fmt.Sprintf("\"blk\": %s,", string(blkjson))
	}
	var jsonString string
	if len(blks) > 0 {

		jsonString = string(jsonConfig[:len(jsonConfig)-1]) + "," + blks[:len(blks)-1] + "}"

	} else {
		jsonString = string(jsonConfig)
	}

	return jsonString, nil

}