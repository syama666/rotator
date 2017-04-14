package rotator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testPath = "test_size.log"
)

func cleanup() {
	os.Remove(testPath)
	os.Remove(testPath + ".1")
	os.Remove(testPath + ".2")
	os.Remove(testPath + ".3")
}

func TestSizeNormalOutput(t *testing.T) {

	cleanup()
	defer cleanup()

	rotator := NewSizeRotator(testPath)
	defer rotator.Close()

	rotator.WriteString("SAMPLE LOG")

	file, err := os.OpenFile(testPath, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	b := make([]byte, 10)
	file.Read(b)
	assert.Equal(t, "SAMPLE LOG", string(b))

	rotator.WriteString("|NEXT LOG")
	rotator.WriteString("|LAST LOG")

	b = make([]byte, 28)
	file.ReadAt(b, 0)

	assert.Equal(t, "SAMPLE LOG|NEXT LOG|LAST LOG", string(b))

}

func TestSizeRotation(t *testing.T) {

	cleanup()
	defer cleanup()

	rotator := NewSizeRotator(testPath)
	rotator.RotationSize = 10
	defer rotator.Close()

	rotator.WriteString("0123456789")
	// it should not be rotated
	stat, _ := os.Lstat(testPath + ".1")
	assert.Nil(t, stat)

	// it should be rotated
	rotator.WriteString("0123456789")
	stat, _ = os.Lstat(testPath)
	assert.NotNil(t, stat)
	assert.Equal(t, stat.Size(), int64(10))

	stat, _ = os.Lstat(testPath + ".1")
	assert.NotNil(t, stat)
	assert.Equal(t, stat.Size(), int64(10))

}

func TestSizeAppendExist(t *testing.T) {

	cleanup()
	defer cleanup()

	file, _ := os.OpenFile(testPath, os.O_WRONLY|os.O_CREATE, 0644)
	file.WriteString("01234") // size should be 5
	file.Close()

	rotator := NewSizeRotator(testPath)
	rotator.RotationSize = 10
	_, err := rotator.WriteString("56789012")
	assert.Nil(t, err)

	stat, _ := os.Lstat(testPath)
	assert.NotNil(t, stat)
	assert.Equal(t, int64(8), stat.Size())

	stat, _ = os.Lstat(testPath + ".1")
	assert.NotNil(t, stat)
	assert.Equal(t, int64(5), stat.Size())

}

func TestSizeMaxRotation(t *testing.T) {

	cleanup()
	defer cleanup()

	rotator := NewSizeRotator(testPath)
	rotator.RotationSize = 10
	rotator.MaxRotation = 3
	defer rotator.Close()

	rotator.WriteString("0123456789")
	stat, _ := os.Lstat(testPath + ".1")
	assert.Nil(t, stat)

	rotator.WriteString("0123456789")
	rotator.WriteString("0123456789")
	rotator.WriteString("0123456789")

	stat, _ = os.Lstat(testPath + ".1")
	assert.NotNil(t, stat)
	assert.Equal(t, stat.Size(), int64(10))

	stat, _ = os.Lstat(testPath + ".2")
	assert.NotNil(t, stat)
	assert.Equal(t, stat.Size(), int64(10))

	stat, _ = os.Lstat(testPath + ".3")
	assert.NotNil(t, stat)
	assert.Equal(t, stat.Size(), int64(10))

	// it should fail rotation
	_, err := rotator.WriteString("0123456789")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "rotation count has been exceeded")
}
