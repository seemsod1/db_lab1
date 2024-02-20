package utils

import (
	"errors"
	"fmt"
	"github.com/seemsod1/db_lab1/internal/driver"
	myErr "github.com/seemsod1/db_lab1/internal/error"
	"github.com/seemsod1/db_lab1/internal/models"
	"log"
	"os"
	"sort"
	"strings"
)

func AddNode(record models.Order, file *os.File, pos int64, prevRecordPos ...int64) bool {
	var prev int64
	if len(prevRecordPos) == 0 {
		prev = -1
		record.Prev = -1
	} else {
		prev = prevRecordPos[0]
		record.Prev = prev
	}
	if !driver.WriteModel(file, record, pos) {
		log.Println("Error: write failed")
		return false
	}

	if prev == -1 {
		return true
	}

	return SetNextPtr(file, prev, pos)
}
func SetNextPtr(file *os.File, recordPos int64, nextRecordPos int64) bool {
	var tmp models.Order

	if !driver.ReadModel(file, &tmp, recordPos) {
		log.Println("Error: Read failed")
		return false
	}

	tmp.Next = nextRecordPos

	if !driver.WriteModel(file, tmp, recordPos) {
		log.Println("Error: Write failed")
		return false
	}
	return true
}
func DeleteOrderNode(file *os.File, recordPos int64, prevRecordPos int64) bool {
	var tmp models.Order
	if !driver.ReadModel(file, &tmp, recordPos) {
		log.Println("Error: read failed")
		return false
	}
	//check if it's first node
	if prevRecordPos == -1 {
		//check if only one node
		if tmp.Next != -1 {
			var next models.Order
			if !driver.ReadModel(file, &next, tmp.Next) {
				log.Println("Error: read failed")
				return false
			}
			next.Prev = -1
			if !driver.WriteModel(file, next, tmp.Next) {
				log.Println("Error: write failed")
				return false
			}
		}
		//delete node
		tmp.Deleted = true
		if !driver.WriteModel(file, tmp, recordPos) {
			log.Println("Error: write failed")
			return false
		}

		return true
	}
	if tmp.Next == -1 {
		//delete node
		tmp.Deleted = true
		if !driver.WriteModel(file, tmp, recordPos) {
			log.Println("Error: write failed")
			return false
		}
		//set previous node's next pointer to -1
		var prev models.Order
		if !driver.ReadModel(file, &prev, prevRecordPos) {
			log.Println("Error: read failed")
			return false
		}
		prev.Next = -1
		if !driver.WriteModel(file, prev, prevRecordPos) {
			log.Println("Error: write failed")
			return false
		}
		return true
	}

	var prev models.Order
	if !driver.ReadModel(file, &prev, prevRecordPos) {
		log.Println("Error: read failed")
		return false
	}

	prev.Next = tmp.Next
	if !driver.WriteModel(file, prev, prevRecordPos) {
		log.Println("Error: write failed")
		return false
	}
	tmp.Deleted = true
	if !driver.WriteModel(file, tmp, recordPos) {
		log.Println("Error: write failed")
		return false
	}

	var next models.Order
	if !driver.ReadModel(file, &next, tmp.Next) {
		log.Println("Error: read failed")
		return false
	}
	next.Prev = prevRecordPos
	if !driver.WriteModel(file, next, tmp.Next) {
		log.Println("Error: write failed")
		return false
	}

	return true
}

func GetAddressById(id uint32, ind []driver.IndexTable) int64 {
	for _, v := range ind {
		if v.Id == id {
			return v.Pos
		}
	}
	return -1
}
func GetIdByAddress(pos int64, ind []driver.IndexTable) uint32 {
	for _, v := range ind {
		if v.Pos == pos {
			return v.Id
		}
	}
	return 0
}
func FindLastNode(file *os.File, recordPos int64, model interface{}) int64 {
	var tmp int64

	for {
		if !driver.ReadModel(file, model, recordPos) {
			return -1
		}
		switch modelTmp := model.(type) {
		case *models.SHeader:
			tmp = modelTmp.Next
		case *models.Order:
			tmp = modelTmp.Next
		default:
			return -1
		}
		if tmp == -1 {
			break
		}
		recordPos = tmp
	}
	return recordPos
}

func AddGarbageNode(file *os.File, garbage *models.SHeader, readPos int64, data any) bool {
	garbage.Next = readPos
	if !driver.WriteModel(file, garbage, garbage.Pos) {
		log.Println("Error: write failed")
		return false
	}
	if !driver.WriteModel(file, data, readPos) {
		log.Println("Error: write failed")
		return false
	}
	garbage.Prev = garbage.Pos
	garbage.Pos = readPos
	garbage.Next = -1
	if !driver.WriteModel(file, garbage, readPos) {
		log.Println("Error: write failed")
		return false
	}
	return true
}
func DeleteGarbageNode(file *os.File, garbage *models.SHeader) *models.SHeader {

	if garbage.Prev == -1 {
		garbage.Next = -1
		if !driver.WriteModel(file, garbage, garbage.Pos) {
			log.Println("Error: write failed")
			return nil
		}
		return garbage
	}

	var prev models.SHeader
	if !driver.ReadModel(file, &prev, garbage.Prev) {
		fmt.Println("Error: read failed")
		return nil
	}

	prev.Next = garbage.Next
	if !driver.WriteModel(file, prev, garbage.Prev) {
		fmt.Println("Error: write failed")
		return nil
	}
	return &prev
}
func createGarbageIndexTable(file *os.File, pos int64) []driver.IndexTable {
	var indices []driver.IndexTable
	i := uint32(0)
	var garbage models.SHeader
	readPos := pos
	for readPos != -1 {
		if !driver.ReadModel(file, &garbage, readPos) {
			return nil
		}
		indices = append(indices, driver.IndexTable{Id: i, Pos: readPos})
		readPos = garbage.Next
		i++
	}
	return indices
}

func ByteArrayToString(bytes []byte) string {
	return strings.TrimRight(string(bytes), "\x00")
}

func CalculateAmountOfNodes(file *os.File, headPos int64, model interface{}) int {
	var amount int

	for headPos != -1 {
		if !driver.ReadModel(file, model, headPos) {
			return -1
		}
		amount++
		switch model := model.(type) {
		case *models.Order:
			headPos = model.Next
		case *models.SHeader:
			headPos = model.Next
		default:
			log.Println("Unsupported model type")
			return -1
		}
	}
	return amount
}

func SortIndicesById(indices []driver.IndexTable) []driver.IndexTable {
	sort.Slice(indices, func(i, j int) bool { return indices[i].Id < indices[j].Id })
	return indices
}
func SortIndicesByPos(indices []driver.IndexTable) []driver.IndexTable {
	sort.Slice(indices, func(i, j int) bool { return indices[i].Pos < indices[j].Pos })
	return indices

}

func CloseFile(firstConfig *driver.FileConfig) (*driver.FileConfig, error) {
	if err := firstConfig.FL.Close(); err != nil {
		return nil, &myErr.Error{Err: myErr.FailedToClose}
	}
	return firstConfig, nil

}
func OptimizeFile(fileConfig *driver.FileConfig, addConfig *driver.FileConfig, isMaster bool) (*driver.FileConfig, error) {
	gabList := createGarbageIndexTable(fileConfig.FL, 0)
	if gabList == nil || len(gabList) < 2 {
		return fileConfig, nil
	}
	gabList = SortIndicesByPos(gabList)
	tmpInd := SortIndicesByPos(fileConfig.Ind)
	trSize := int64(0)
	var err error

	if isMaster {
		trSize = driver.UserSize
	} else {
		trSize = driver.OrderSize
	}
	var pos int64
	if len(tmpInd) == 0 {
		pos = 0
	} else {
		pos = tmpInd[len(tmpInd)-1].Pos
	}
	//delete records which are after last user
	gabList, cutted := cutRecordsBelow(gabList, pos)
	if cutted {
		if isMaster {
			gabList, err = reorderGarbage(fileConfig.FL, gabList, models.User{Deleted: true})
			if err != nil {
				return nil, &myErr.Error{Err: myErr.FailedToOptimize}
			}
		} else {
			gabList, err = reorderGarbage(fileConfig.FL, gabList, models.Order{Deleted: true})
			if err != nil {
				return nil, &myErr.Error{Err: myErr.FailedToOptimize}
			}
		}

		var garbage models.SHeader
		if !driver.ReadModel(fileConfig.FL, &garbage, gabList[len(gabList)-1].Pos) {
			return nil, &myErr.Error{Err: myErr.FailedToOptimize}
		}
		garbage.Next = -1
		if !driver.WriteModel(fileConfig.FL, garbage, gabList[len(gabList)-1].Pos) {
			return nil, &myErr.Error{Err: myErr.FailedToOptimize}
		}
		fileConfig.GarbageNode = &garbage

		if err := fileConfig.FL.Truncate(pos + trSize); err != nil {
			return nil, &myErr.Error{Err: myErr.FailedToOptimize}
		}

		fileConfig.Pos = pos + trSize

	}
	if len(gabList) < 2 {
		return fileConfig, nil
	}
	gabList = gabList[1:]
	log.Println("Checking fragmentation percentage")
	if calculateFragmentationPercentage(fileConfig.Ind, isMaster) > driver.FragmentationThreshold*100 {
		fmt.Println("Need to defragment")
		//defragment
		if tmpInd, err = defragmentFile(fileConfig.FL, fileConfig.Ind, gabList, addConfig, isMaster); err != nil {
			return nil, &myErr.Error{Err: myErr.FailedToOptimize}
		}
		if err := fileConfig.FL.Truncate(tmpInd[len(tmpInd)-1].Pos + trSize); err != nil {
			return nil, &myErr.Error{Err: myErr.FailedToOptimize}
		}
		if !driver.WriteModel(fileConfig.FL, &models.SHeader{Prev: -1, Pos: 0, Next: -1}, 0) {
			return nil, &myErr.Error{Err: myErr.FailedToOptimize}
		}
		fileConfig.GarbageNode = &models.SHeader{Prev: -1, Pos: 0, Next: -1}
		fileConfig.Ind = tmpInd
		fileConfig.Pos = tmpInd[len(tmpInd)-1].Pos + trSize
		return fileConfig, nil
	}
	log.Println("No need to defragment")
	tmpInd = SortIndicesById(tmpInd)
	fileConfig.Ind = tmpInd
	return fileConfig, nil
}
func reorderGarbage(file *os.File, indices []driver.IndexTable, model any) ([]driver.IndexTable, error) {
	var garbage models.SHeader
	var ind []driver.IndexTable
	readPos := int64(0)
	if !driver.ReadModel(file, &garbage, readPos) {
		return nil, fmt.Errorf("Error: read failed")
	}
	ind = append(ind, indices[0])
	indices = append(indices[:0], indices[1:]...)
	for _, i := range indices {
		readPos = i.Pos
		if !AddGarbageNode(file, &garbage, readPos, model) {
			return nil, fmt.Errorf("Error: add garbage node failed")
		}
		ind = append(ind, driver.IndexTable{Id: i.Id, Pos: readPos})
	}
	return ind, nil
}
func cutRecordsBelow(slice []driver.IndexTable, lastUserPos int64) ([]driver.IndexTable, bool) {
	size := len(slice)
	slice = SortIndicesByPos(slice)

	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i].Pos > lastUserPos {
			slice = append(slice[:i], slice[i+1:]...)
		} else {
			break
		}
	}

	return slice, size != len(slice)
}
func cutRecordsAbove(slice []driver.IndexTable, lastUserPos int64) ([]driver.IndexTable, bool) {
	size := len(slice)
	slice = SortIndicesByPos(slice)

	for i := 0; i < len(slice); i++ {
		if slice[i].Pos < lastUserPos {
			slice = append(slice[:i], slice[i+1:]...)
		} else {
			break
		}
	}

	return slice, size != len(slice)
}

func calculateFragmentationPercentage(table []driver.IndexTable, isMaster bool) float64 {
	table = SortIndicesByPos(table)

	var size int64
	if isMaster {
		size = driver.UserSize
	} else {
		size = driver.OrderSize
	}

	fullFileSize := table[len(table)-1].Pos

	emptySpace := fullFileSize - int64(len(table))*size

	return float64(emptySpace) / float64(fullFileSize) * 100
}
func defragmentFile(file *os.File, ind []driver.IndexTable, gab []driver.IndexTable, addConfig *driver.FileConfig, isMaster bool) ([]driver.IndexTable, error) {
	ind = SortIndicesByPos(ind)
	gab = SortIndicesByPos(gab)
	if isMaster {
		afterOpt := int64(len(ind)) * driver.UserSize
		for ind[len(ind)-1].Pos != afterOpt {
			var user models.User
			if !driver.ReadModel(file, &user, ind[len(ind)-1].Pos) {
				return nil, &myErr.Error{Err: myErr.FailedToOptimize}
			}
			ind = removeByPos(ind[len(ind)-1].Pos, ind)
			if !driver.WriteModel(file, user, gab[0].Pos) {
				return nil, &myErr.Error{Err: myErr.FailedToOptimize}
			}
			ind = append(ind, driver.IndexTable{Id: user.ID, Pos: gab[0].Pos})
			ind = SortIndicesByPos(ind)
			gab = removeByPos(gab[0].Pos, gab)

		}

		return ind, nil

	} else {
		//reorder slave file
		afterOpt := int64(len(ind)) * driver.OrderSize
		var optimized bool
		//first reorder tails
		tails := make([]driver.IndexTable, 0)
		heads := make([]driver.IndexTable, 0)
		for _, v := range ind {
			var order models.Order
			if !driver.ReadModel(file, &order, v.Pos) {
				return nil, &myErr.Error{Err: myErr.FailedToOptimize}
			}
			if order.Next == -1 {
				tails = append(tails, v)
			}
			if order.Prev == -1 {
				heads = append(heads, v)
			}
		}
		userIdHeadPos := make([]driver.IndexTable, 0)
		for _, v := range addConfig.Ind {
			var user models.User
			if !driver.ReadModel(addConfig.FL, &user, v.Pos) {
				return nil, &myErr.Error{Err: myErr.FailedToOptimize}
			}
			if user.FirstOrder != -1 {
				userIdHeadPos = append(userIdHeadPos, driver.IndexTable{Id: user.ID, Pos: user.FirstOrder})
			}
		}
		tails = SortIndicesByPos(tails)
		var cutted bool
		_, cutted = cutRecordsBelow(tails, afterOpt)
		if cutted {
			tails, _ = cutRecordsAbove(tails, afterOpt)
			for i := len(tails) - 1; i >= 0; i-- {
				v := tails[i]
				var order models.Order
				if !driver.ReadModel(file, &order, v.Pos) {
					return nil, &myErr.Error{Err: myErr.FailedToOptimize}
				}
				ind = removeByPos(v.Pos, ind)
				id := order.OrderId
				if !driver.WriteModel(file, order, gab[0].Pos) {
					return nil, &myErr.Error{Err: myErr.FailedToOptimize}
				}
				prevPos := order.Prev
				if prevPos != -1 {
					if !driver.ReadModel(file, &order, prevPos) {
						return nil, &myErr.Error{Err: myErr.FailedToOptimize}
					}
					order.Next = gab[0].Pos
					if !driver.WriteModel(file, order, prevPos) {
						return nil, &myErr.Error{Err: myErr.FailedToOptimize}
					}
				} else {
					//this is head       v.Pos
					//change headpos in user file
					userPos := GetAddressById(order.UserId, addConfig.Ind)
					var user models.User
					if !driver.ReadModel(addConfig.FL, &user, userPos) {
						return nil, &myErr.Error{Err: myErr.FailedToOptimize}
					}
					user.FirstOrder = gab[0].Pos
					if !driver.WriteModel(addConfig.FL, user, userPos) {
						return nil, &myErr.Error{Err: myErr.FailedToOptimize}
					}
					//delete from heads
					heads = removeByPos(v.Pos, heads)
				}
				ind = append(ind, driver.IndexTable{Id: id, Pos: gab[0].Pos})
				ind = SortIndicesByPos(ind)
				gab = removeByPos(gab[0].Pos, gab)
				tails = removeByPos(v.Pos, tails)
				if ind[len(ind)-1].Pos == afterOpt {
					optimized = true
					break
				}
			}
		}
		if !optimized {
			//then reorder heads

			heads = SortIndicesByPos(heads)
			_, cutted = cutRecordsBelow(heads, afterOpt)
			if cutted {
				heads, _ = cutRecordsAbove(heads, afterOpt)
				for i := len(heads) - 1; i >= 0; i-- {
					v := heads[i]
					var order models.Order
					if !driver.ReadModel(file, &order, v.Pos) {
						return nil, &myErr.Error{Err: myErr.FailedToOptimize}
					}
					id := order.OrderId
					ind = removeByPos(v.Pos, ind)
					if !driver.WriteModel(file, order, gab[0].Pos) {
						return nil, &myErr.Error{Err: myErr.FailedToOptimize}
					}
					nextPos := order.Next

					if !driver.ReadModel(file, &order, nextPos) {
						return nil, &myErr.Error{Err: myErr.FailedToOptimize}
					}
					order.Prev = gab[0].Pos
					if !driver.WriteModel(file, order, nextPos) {
						return nil, &myErr.Error{Err: myErr.FailedToOptimize}
					}

					//change headpos in user file
					var user models.User
					userPos := GetAddressById(order.UserId, addConfig.Ind)
					if !driver.ReadModel(addConfig.FL, &user, userPos) {
						return nil, &myErr.Error{Err: myErr.FailedToOptimize}
					}
					user.FirstOrder = gab[0].Pos
					if !driver.WriteModel(addConfig.FL, user, userPos) {
						return nil, &myErr.Error{Err: myErr.FailedToOptimize}
					}

					ind = append(ind, driver.IndexTable{Id: id, Pos: gab[0].Pos})
					ind = SortIndicesByPos(ind)
					gab = removeByPos(gab[0].Pos, gab)
					heads = removeByPos(v.Pos, heads)
					if ind[len(ind)-1].Pos == afterOpt {
						optimized = true
						break
					}
				}
			}

		} else {
			return ind, nil
		}

		if !optimized {
			//then reorder the rest
			for ind[len(ind)-1].Pos != afterOpt {
				var order models.Order
				if !driver.ReadModel(file, &order, ind[len(ind)-1].Pos) {
					return nil, &myErr.Error{Err: myErr.FailedToOptimize}
				}
				id := order.OrderId
				ind = removeByPos(ind[len(ind)-1].Pos, ind)
				if !driver.WriteModel(file, order, gab[0].Pos) {
					return nil, &myErr.Error{Err: myErr.FailedToOptimize}
				}
				var nextOrder models.Order
				nextPos := order.Next

				if !driver.ReadModel(file, &nextOrder, nextPos) {
					return nil, &myErr.Error{Err: myErr.FailedToOptimize}
				}
				nextOrder.Prev = gab[0].Pos
				if !driver.WriteModel(file, nextOrder, nextPos) {
					return nil, &myErr.Error{Err: myErr.FailedToOptimize}
				}

				var prevOrder models.Order
				prevPos := order.Prev

				if !driver.ReadModel(file, &prevOrder, prevPos) {
					return nil, &myErr.Error{Err: myErr.FailedToOptimize}
				}
				prevOrder.Next = gab[0].Pos
				if !driver.WriteModel(file, prevOrder, prevPos) {
					return nil, &myErr.Error{Err: myErr.FailedToOptimize}
				}

				ind = append(ind, driver.IndexTable{Id: id, Pos: gab[0].Pos})
				ind = SortIndicesByPos(ind)
				gab = removeByPos(gab[0].Pos, gab)

			}
			optimized = true
		} else {
			return ind, nil
		}

		if !optimized {
			return nil, &myErr.Error{Err: myErr.FailedToOptimize}
		} else {
			return ind, nil
		}
	}
}
func removeByPos(pos int64, indices []driver.IndexTable) []driver.IndexTable {
	for i, v := range indices {
		if v.Pos == pos {
			if len(indices) == 1 {
				return nil
			} else {
				indices = append(indices[:i], indices[i+1:]...)
			}
			break
		}
	}
	return indices
}

func CreateFileConfig(filename string, isMaster bool) (*driver.FileConfig, error) {
	FL, err := os.OpenFile(filename+".fl", os.O_RDWR, 0666)
	var headerNext int64
	var indices []driver.IndexTable
	var garbageNode *models.SHeader
	var size int64
	if isMaster {
		size = driver.UserSize
	} else {
		size = driver.OrderSize

	}
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File doesn't exist, it was created
			FL, err = os.OpenFile(filename+".fl", os.O_RDWR|os.O_CREATE, 0666)
			if isMaster {
				if !driver.WriteModel(FL, models.User{Deleted: true}, 0) {
					return nil, err
				}
				headerNext = size
			} else {
				if !driver.WriteModel(FL, models.Order{Deleted: true}, 0) {
					return nil, err
				}
				headerNext = size
			}
			garbageNode = &models.SHeader{Prev: -1, Pos: 0, Next: -1}
			if !driver.WriteModel(FL, garbageNode, 0) {
				return nil, err
			}

			log.Println("New config created")
		} else {
			// Some other error occurred
			return nil, err
		}
	} else {
		// File was opened
		ind, err := os.OpenFile(filename+".ind", os.O_RDWR, 0666)
		if err != nil {
			log.Fatal(err)
		}

		indices, err = LoadIndices(ind)
		if err != nil {
			return nil, err
		}
		var posInFile int64

		indices = SortIndicesByPos(indices)
		if len(indices) != 0 {
			posInFile = indices[len(indices)-1].Pos + size
		} else {
			posInFile = size
		}
		indices = SortIndicesById(indices)

		var gab models.SHeader
		gabPos := FindLastNode(FL, 0, &gab)
		if !driver.ReadModel(FL, &gab, gabPos) {
			return nil, err
		}
		garbageNode = &gab
		headerNext = posInFile

		log.Println("Config loaded")
	}

	fileConfig := driver.NewFileConfig(FL, headerNext, indices, garbageNode)
	return fileConfig, nil
}
func LoadIndices(indFile *os.File) ([]driver.IndexTable, error) {
	readPos := int64(0)

	var indices []driver.IndexTable
	for {
		var model driver.IndexTable
		if !driver.ReadModel(indFile, &model, readPos) {
			break
		}
		readPos += driver.IndexSize
		indices = append(indices, model)
	}

	return indices, nil
}
func WriteIndices(filename string, indices []driver.IndexTable) {
	FL, err := os.OpenFile(filename+".ind", os.O_RDWR|os.O_CREATE, 0666)
	// File was opened
	if err = FL.Truncate(0); err != nil {
		log.Fatal(err)
	}
	writePos := int64(0)
	for _, v := range indices {
		if !driver.WriteModel(FL, v, writePos) {
			log.Fatal(err)
		}
		writePos += driver.IndexSize
	}
	log.Println("Indices written")
	FL.Close()
}
