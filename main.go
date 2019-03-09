package main

import (
	"fmt"
	"log"

	"github.com/nfnt/resize"
	"image"
	"image/color"
	"image/png"

	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type memoryMap map[string][]int
type weightMap map[string]int

func openPNG(name string) []int {
	imgSlice := make([]int, 0)
	file, err := os.Open(name + ".png")
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	img, err := png.Decode(file)

	if err != nil {
		log.Fatal(err)
	}

	img = resize.Resize(15, 20, img, resize.Bicubic)

	// Распределение каждого пикселя изображения по шкале яркости от 0 до 3
	// и занесение этих значений в массив
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayTransparensy := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			level := (255 - grayTransparensy.Y) / 85
			if level == 3 {
				level--
			}
			imgSlice = append(imgSlice, int(level))
		}
	}

	return imgSlice
}

func openPNGM(name string) [][]int {
	file, err := os.Open(name + ".png")
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	img, err := png.Decode(file)

	if err != nil {
		log.Fatal(err)
	}

	spaces := make([][]byte, 0)

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		tempSpaces := make([]byte, 0)
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayTransparency := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			if grayTransparency.Y > 96 {
				tempSpaces = append(tempSpaces, 1)
			} else {
				tempSpaces = append(tempSpaces, 0)
			}
		}
		spaces = append(spaces, tempSpaces)
	}

	spacesRet := make([]byte, 0)

	for i := range spaces[0] {
		spacesRet = append(spacesRet, 1)
		for j := range spaces {
			if spaces[j][i] == 1 {
				continue
			} else {
				spacesRet[len(spacesRet)-1] = 0
				break
			}
		}
	}

	spacesY := make([][]byte, 0)

	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		tempSpacesY := make([]byte, 0)
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			grayTransparencyY := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			if grayTransparencyY.Y > 175 {
				tempSpacesY = append(tempSpacesY, 1)
			} else {
				tempSpacesY = append(tempSpacesY, 0)
			}
		}
		spacesY = append(spacesY, tempSpacesY)
	}

	spacesRetY := make([]byte, 0)

	for i := range spacesY[0] {
		spacesRetY = append(spacesRetY, 1)
		for j := range spacesY {
			if spacesY[j][i] == 1 {
				continue
			} else {
				spacesRetY[len(spacesRetY)-1] = 0
				break
			}
		}
	}
	posY := make([]int, 0, 2)

	for i := range spacesRetY {
		if i > 0 {
			if spacesRetY[i] != spacesRetY[i-1] {
				posY = append(posY, i)
			}
		}
	}

	var (
		minY int
		maxY int
	)

	switch len(posY) {
	case 0:
		minY = img.Bounds().Min.Y
		maxY = img.Bounds().Max.Y
	case 1:
		minY = posY[0]
		maxY = img.Bounds().Max.Y
	case 2:
		minY = posY[0]
		maxY = posY[1]
	default:
		panic("ERROR")
	}

	imgArr := make([]int, 0)

	for i := range spacesRet {
		if i != 0 {
			if spacesRet[i] != spacesRet[i-1] {
				imgArr = append(imgArr, i)
			}
		}
	}

	var k int
	imgs := make([]*image.NRGBA, 0)
	for i := 0; i < len(imgArr)-1; i += 2 {
		height, width := maxY-minY, imgArr[i+1]-imgArr[i]+1

		// Create a colored image of the given width and height.
		// imgq := image.NewNRGBA(image.Rect(0, 0, width, height))
		imgs = append(imgs, image.NewNRGBA(image.Rect(0, 0, width, height)))

		for y := minY; y < maxY; y++ {
			for x := imgArr[i]; x <= imgArr[i+1]; x++ {
				grayTransparency := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
				if grayTransparency.Y > 245 {
					imgs[k].Set(x-imgArr[i], y-minY, color.NRGBA{
						R: 255,
						G: 255,
						B: 255,
						A: grayTransparency.Y,
					})
				} else {
					imgs[k].Set(x-imgArr[i], y-minY, color.NRGBA{
						R: 0,
						G: 0,
						B: 0,
						A: 255 - grayTransparency.Y,
					})
				}
			}
		}

		k++
	}

	imgSliceRet := make([][]int, 0)

	// Если для обучения нужны символы с нескольких изображений, то запись файлов
	// начинается с индекса, указанного в файле teach/res/lastFileIndex
	text, err := ioutil.ReadFile("teach/res/lastFileIndex")
	var index int

	if err != nil {
		index = 0
	} else {
		tempStr := strings.Trim(string(text), " \n")
		index, err = strconv.Atoi(tempStr)
		if err != nil {
			index = 0
		}
	}

	for i, some := range imgs {
		width := some.Bounds().Max.X - some.Bounds().Min.X
		height := some.Bounds().Max.Y - some.Bounds().Min.Y
		someImage := some.SubImage(image.Rect(0, 0, width, height))

		f, _ := os.Create("teach/file" + fmt.Sprint(i+index) + ".png")
		png.Encode(f, someImage)
		someImage = resize.Resize(15, 20, someImage, resize.Bicubic)
		imgTempSlice := make([]int, 0)

		// Распределение каждого пикселя изображения по шкале яркости от 0 до 3
		// и занесение этих значений в массив
		for y := someImage.Bounds().Min.Y; y < someImage.Bounds().Max.Y; y++ {
			for x := someImage.Bounds().Min.X; x < someImage.Bounds().Max.X; x++ {
				grayTransparensy := color.GrayModel.Convert(someImage.At(x, y)).(color.Gray)
				level := (255 - grayTransparensy.Y) / 85
				if level == 3 {
					level--
				}
				imgTempSlice = append(imgTempSlice, int(level))
			}
		}

		imgSliceRet = append(imgSliceRet, imgTempSlice)
	}

	ioutil.WriteFile("teach/res/lastFileIndex", []byte(strconv.Itoa(len(imgs)+index)), 0644)
	return imgSliceRet
}

func openMemoryFile(name string) []int {
	memorySlice := make([]int, 0)
	file, _ := ioutil.ReadFile("src/" + name)
	fileString := strings.Replace(string(file), "\n", "", -1)

	for _, k := range strings.Split(fileString, " ") {
		memoryInt, err := strconv.Atoi(k)

		if err != nil {
			continue
		}

		memorySlice = append(memorySlice, memoryInt)
	}

	return memorySlice
}

func writeMemoryFile(name string, memorySlice []int) {
	var stringToMemory string

	for i := range memorySlice {
		stringToMemory += strconv.Itoa(memorySlice[i]) + " "
	}

	stringToMemory = strings.TrimSuffix(stringToMemory, " ")
	ioutil.WriteFile("src/"+name, []byte(stringToMemory), 0644)
	return
}

func recognize(image []int, memoryTable memoryMap) string {
	weightTable := make(weightMap)

	// Сравнение пикселей исходной картинки с данными из памяти машины
	// Если пиксель картинки и пиксель из памяти не равен нулю то
	// значения складываются
	for i, memorySlice := range memoryTable {
		for j, memoryInt := range memorySlice {
			if memoryInt != 0 && image[j] != 0 {
				weightTable[i] += memoryInt
			}
		}
	}

	// Проверка на то, что сеть обучена
	var allWeightZero bool = true

	for _, weight := range weightTable {
		if weight != 0 {
			allWeightZero = false
			break
		}
	}

	var result string

	if allWeightZero {
		result = "5"
	} else {

		// Нахождение максимального веса(ов)
		maxWeight := make([]int, 0)
		maxWeightName := make([]string, 0)
		maxWeight = append(maxWeight, 0)
		maxWeightName = append(maxWeightName, "")
		for i, weight := range weightTable {
			if weight > maxWeight[len(maxWeight)-1] {
				maxWeight[len(maxWeight)-1] = weight
				maxWeightName[len(maxWeightName)-1] = i
			} else if weight == maxWeight[len(maxWeight)-1] {
				maxWeight = append(maxWeight, weight)
				maxWeightName = append(maxWeightName, i)
			}
		}

		// Если есть несколько букв у которых веса совпадают то проводится проверка
		// на веса в случаях, когда степень прозрачности на данной картинке равна нулю
		// а в данных не равна нулю. Результатом считается символ с наименьшем значением веса
		if len(maxWeight) > 1 {
			tempMap := make(map[string]int)

			for _, name := range maxWeightName {
				for i, memoryInt := range memoryTable[name] {
					if image[i] == 0 {
						tempMap[name] += memoryInt
					}
				}
			}

			minOther := 0
			minOtherName := ""

			for i := range tempMap {
				minOther = tempMap[i]
				minOtherName = i
				break
			}
			fmt.Println(minOtherName)

			for i := range tempMap {
				if tempMap[i] < minOther {
					minOther = tempMap[i]
					minOtherName = i
				}
			}

			result = minOtherName
		} else {
			result = maxWeightName[0]
		}
	}

	return result
}

func teach(result string, want string, memoryTable memoryMap, fileName string) {
	image := openPNG(fileName)
	memory1 := memoryTable[result]
	memory2 := memoryTable[want]

	for i := range image {
		if memory1[i] <= memory2[i] && image[i] != 0 {
			memory1[i] -= image[i]
		}

		memory2[i] += image[i]
	}

	writeMemoryFile(result, memory1)
	writeMemoryFile(want, memory2)
}

func main() {
	// n, _ := ioutil.ReadFile("name.txt")
	// names := strings.Trim(string(n), "\n ")
	// nameTable := strings.Split(string(names), " ")
	nameTable := make([]string, 0)
	folder, _ := ioutil.ReadDir("teach")

	for _, file := range folder {
		if strings.HasSuffix(file.Name(), ".png") {
			nameTable = append(nameTable, strings.TrimSuffix(file.Name(), ".png"))
		}
	}

	memoryTable := make(memoryMap)

	for _, name := range nameTable {
		f, _ := ioutil.ReadFile("teach/" + name + ".txt")
		file := strings.Trim(string(f), " \n")
		memoryTable[file] = openMemoryFile(file)
	}

	if os.Args[1] == "-t" {
		var teachIndex int
		var isEdited bool

		for {
			if teachIndex == len(nameTable) {
				break
			}

			name := nameTable[teachIndex]
			log.Printf("OPENING %s\n", name)
			f, _ := ioutil.ReadFile("teach/" + name + ".txt")
			want := strings.Trim(string(f), " \n")
			image := openPNG("teach/" + name)
			result := recognize(image, memoryTable)

			if result == want {
				log.Printf("YES WANT: %s\n", result)

				if isEdited {
					isEdited = false
					teachIndex = 0
				} else {
					teachIndex++
				}

			} else {
				isEdited = true
				log.Printf("NO WANT: %s HAVE: %s\n", want, result)
				teach(result, want, memoryTable, "teach/"+name)
			}
		}
	} else if os.Args[1] == "-m" {
		imgSlice := openPNGM(os.Args[2])

		for _, image := range imgSlice {
			fmt.Print(recognize(image, memoryTable))
		}
		fmt.Println("")
	} else {
		fileName := os.Args[1]
		image := openPNG(fileName)
		fmt.Println(image)
		result := recognize(image, memoryTable)
		fmt.Printf("Is this %s? y/n > ", result)
		var vote string
		fmt.Scan(&vote)

		if vote == "y" {
			return
		} else {
			fmt.Printf("What is this? > ")
			var want string
			fmt.Scan(&want)
			teach(result, want, memoryTable, fileName)
		}
	}
}
