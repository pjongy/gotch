package vision

// A simple dataset structure shared by various computer vision datasets.

import (
	"log"
	"math/rand"
	"time"

	ts "github.com/sugarme/gotch/tensor"
)

type Dataset struct {
	TrainImages ts.Tensor
	TrainLabels ts.Tensor
	TestImages  ts.Tensor
	TestLabels  ts.Tensor
	Labels      int64
}

// Dataset Methods:
//=================

// TrainIter creates an iterator of Iter type for train images and labels
func (ds Dataset) TrainIter(batchSize int64) (retVal ts.Iter2) {
	return ts.MustNewIter2(ds.TrainImages, ds.TrainLabels, batchSize)

}

// TestIter creates an iterator of Iter type for test images and labels
func (ds Dataset) TestIter(batchSize int64) (retVal ts.Iter2) {
	return ts.MustNewIter2(ds.TestImages, ds.TestLabels, batchSize)
}

// RandomFlip randomly applies horizontal flips
// This expects a 4 dimension NCHW tensor and returns a tensor with
// an identical shape.
func RandomFlip(t ts.Tensor) (retVal ts.Tensor) {

	size := t.MustSize()

	if len(size) != 4 {
		log.Fatalf("Unexpected shape for tensor %v\n", size)
	}

	output, err := t.ZerosLike(false)
	if err != nil {
		panic(err)
	}

	for batchIdx := 0; batchIdx < int(size[0]); batchIdx++ {
		outputView := output.Idx(ts.NewSelect(int64(batchIdx)))
		tView := t.Idx(ts.NewSelect(int64(batchIdx)))

		var src ts.Tensor
		if rand.Float64() == 1.0 {
			src = tView
		} else {
			src = tView.MustFlip([]int64{2})
		}

		tView.MustDrop()
		outputView.Copy_(src)
		src.MustDrop()
		outputView.MustDrop()
	}

	return output
}

// Pad the image using reflections and take some random crops.
// This expects a 4 dimension NCHW tensor and returns a tensor with
// an identical shape.
func RandomCrop(t ts.Tensor, pad int64) (retVal ts.Tensor) {

	size := t.MustSize()

	if len(size) < 4 {
		log.Fatalf("Unexpected shape (%v) for tensor %v\n", size, t)
	}

	szH := size[2]
	szW := size[3]
	padded := t.MustReflectionPad2d([]int64{pad, pad, pad, pad})
	output, err := t.ZerosLike(false)
	if err != nil {
		log.Fatal(err)
	}

	for bidx := 0; bidx < int(size[0]); bidx++ {
		idx := ts.NewSelect(int64(bidx))
		outputView := output.Idx(idx)

		rand.Seed(time.Now().UnixNano())
		startW := rand.Intn(int(2 * pad))
		startH := rand.Intn(int(2 * pad))

		var srcIdx []ts.TensorIndexer
		nIdx := ts.NewSelect(int64(bidx))
		cIdx := ts.NewSelect(int64(-1))
		hIdx := ts.NewNarrow(int64(startH), int64(startH)+szH)
		wIdx := ts.NewNarrow(int64(startW), int64(startW)+szW)
		srcIdx = append(srcIdx, nIdx, cIdx, hIdx, wIdx)
		src := padded.Idx(srcIdx)
		outputView.Copy_(src)
		src.MustDrop()
		outputView.MustDrop()
	}

	padded.MustDrop()

	return output
}

// Applies cutout: randomly remove some square areas in the original images.
// https://arxiv.org/abs/1708.04552
func RandomCutout(t ts.Tensor, sz int64) (retVal ts.Tensor) {

	size := t.MustSize()

	if len(size) != 4 || sz > size[2] || sz > size[3] {
		log.Fatalf("Unexpected shape (%v) for tensor %v\n", size, t)
	}

	output, err := t.ZerosLike(false)
	if err != nil {
		log.Fatal(err)
	}

	output.Copy_(t)

	for bidx := 0; bidx < int(size[0]); bidx++ {

		rand.Seed(time.Now().UnixNano())
		startH := rand.Intn(int(size[2] - sz + 1))
		startW := rand.Intn(int(size[3] - sz + 1))

		var srcIdx []ts.TensorIndexer
		nIdx := ts.NewSelect(int64(bidx))
		cIdx := ts.NewSelect(int64(-1))
		hIdx := ts.NewNarrow(int64(startH), int64(startH)+sz)
		wIdx := ts.NewNarrow(int64(startW), int64(startW)+sz)
		srcIdx = append(srcIdx, nIdx, cIdx, hIdx, wIdx)

		tmp := output.Idx(srcIdx)
		tmp.Fill_(ts.FloatScalar(0.0))
		tmp.MustDrop()
	}

	return output
}

func Augmentation(t ts.Tensor, flip bool, crop int64, cutout int64) (retVal ts.Tensor) {

	tclone := t.MustShallowClone()

	var flipTs ts.Tensor
	if flip {
		flipTs = RandomFlip(tclone)
	} else {
		flipTs = tclone
	}

	tclone.MustDrop()
	return flipTs

	// var cropTs ts.Tensor
	// if crop > 0 {
	// cropTs = RandomCrop(flipTs, crop)
	// } else {
	// cropTs = flipTs
	// }
	//
	// if cutout > 0 {
	// retVal = RandomCutout(cropTs, cutout)
	// } else {
	// retVal = cropTs
	// }

	// tclone.MustDrop()
	// flipTs.MustDrop()
	// // cropTs.MustDrop()
	//
	// return retVal
}
