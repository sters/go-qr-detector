package main

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"github.com/jessevdk/go-flags"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/multi/qrcode/detector"
	"github.com/morikuni/failure"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

var errInvalidPoints = errors.New("invalid points")

type opt struct {
	File    string `description:"file path for target image file" long:"file"    required:"true" short:"f" value-name:"<file>"`
	Verbose bool   `description:"verbose mode"                    long:"verbose" short:"v"`
}

func extendLine(p1 gozxing.ResultPoint, p2 gozxing.ResultPoint, offset float64) (gozxing.ResultPoint, gozxing.ResultPoint) {
	vecX := p2.GetX() - p1.GetX()
	vecY := p2.GetY() - p1.GetY()

	length := math.Sqrt(vecX*vecX + vecY*vecY)
	if length == 0 {
		return p1, p2
	}

	// 単位ベクトルを計算（方向）
	unitVecX := vecX / length
	unitVecY := vecY / length

	return gozxing.NewResultPoint(
			p1.GetX()-offset*unitVecX,
			p1.GetY()-offset*unitVecY,
		),
		gozxing.NewResultPoint(
			p2.GetX()+offset*unitVecX,
			p2.GetY()+offset*unitVecY,
		)
}

func detect(opts opt) error {
	file, err := os.Open(opts.File)
	if err != nil {
		return failure.Wrap(err)
	}
	defer func() { _ = file.Close() }()

	img, _, err := image.Decode(file)
	if err != nil {
		return failure.Wrap(err)
	}

	zxingImg, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return failure.Wrap(err)
	}

	matrix, e := zxingImg.GetBlackMatrix()
	if e != nil {
		return failure.Wrap(err)
	}

	d := detector.NewMultiDetector(matrix)
	results, err := d.DetectMulti(map[gozxing.DecodeHintType]any{
		gozxing.DecodeHintType_TRY_HARDER: true,
	})
	if err != nil {
		if opts.Verbose {
			_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		}

		return failure.Wrap(err)
	}

	dc := gg.NewContextForImage(img)

	errs := make([]error, 0, len(results))
	for _, result := range results {
		pp := result.GetPoints()
		if len(pp) != 3 {
			errs = append(errs, fmt.Errorf("%w: %d", errInvalidPoints, len(pp)))

			continue
		}

		pp = append(pp, gozxing.NewResultPoint(
			pp[0].GetX()-pp[1].GetX()+pp[2].GetX(),
			pp[0].GetY()-pp[1].GetY()+pp[2].GetY(),
		))

		offset := float64(result.GetBits().GetWidth()+result.GetBits().GetHeight()) / 4

		p00, p10 := extendLine(pp[0], pp[1], offset)
		dc.DrawLine(p00.GetX(), p00.GetY(), p10.GetX(), p10.GetY())
		p11, p20 := extendLine(pp[1], pp[2], offset)
		dc.DrawLine(p11.GetX(), p11.GetY(), p20.GetX(), p20.GetY())
		p21, p30 := extendLine(pp[2], pp[3], offset)
		dc.DrawLine(p21.GetX(), p21.GetY(), p30.GetX(), p30.GetY())
		p31, p01 := extendLine(pp[3], pp[0], offset)
		dc.DrawLine(p31.GetX(), p31.GetY(), p01.GetX(), p01.GetY())
		dc.SetRGB255(255, 0, 0)
		dc.SetLineWidth(2)
		dc.Stroke()

		if opts.Verbose {
			fmt.Printf(
				"(%.1f, %.1f), (%.1f, %.1f), (%.1f, %.1f), (%.1f, %.1f)\n",
				pp[0].GetX(), pp[0].GetY(),
				pp[1].GetX(), pp[1].GetY(),
				pp[2].GetX(), pp[2].GetY(),
				pp[3].GetX(), pp[3].GetY(),
			)
		}
	}
	if len(errs) != 0 {
		return failure.Wrap(errors.Join(errs...))
	}

	outfile, err := os.Create(strings.Replace(opts.File, filepath.Ext(opts.File), "_detected.png", 1))
	if err != nil {
		return failure.Wrap(err)
	}
	defer func() { _ = outfile.Close() }()

	if err := png.Encode(outfile, dc.Image()); err != nil {
		return failure.Wrap(err)
	}

	return nil
}

func main() {
	var opts opt
	if _, err := flags.Parse(&opts); err != nil {
		flags.WroteHelp(err)
		os.Exit(1)
	}

	err := detect(opts)
	if err != nil {
		if opts.Verbose {
			_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		}

		os.Exit(1)
	}
}
