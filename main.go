package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

type Kriteria struct {
	Kode          string                 `json:"kode"`
	Nama          string                 `json:"nama"`
	Bobot         float64                `json:"bobot"`
	Atribut       bool                   `json:"atribut"`
	NilaiKriteria map[string]interface{} `json:"nilai_kriteria"`
}

type Alternatif struct {
	NamaMahasiswa string
	NilaiKriteria []float64
	TotalNilai    float64
}

func (alternatif *Alternatif) UbahNilaiKriteria(index int, nilaiBaru float64) {
	alternatif.NilaiKriteria[index] = nilaiBaru
}

func (alternatif *Alternatif) UbahTotalNilai(nilaiBaru float64) {
	alternatif.TotalNilai = nilaiBaru
}

func main() {
	// Kriteria
	// Mengambil kriteria dari file kriteria.json
	fmt.Println("Mengambil data kriteria...")
	jsonFile, err := os.Open("data/kriteria.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	// Mapping data json keriteria kedalam struct kriteria
	var kriteria []Kriteria
	keriteraJsonFile, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(keriteraJsonFile, &kriteria)

	// Menampilkan table kriteria
	tableRenderer := tablewriter.NewWriter(os.Stdout)
	tableRenderer.SetHeader([]string{"Nama Kriteria", "Kode", "Bobot", "Atribut"})
	for _, val := range kriteria {
		ketAtribut := "Cost"
		if val.Atribut {
			ketAtribut = "Benefit"
		}
		tableRenderer.Append([]string{
			val.Nama,
			val.Kode,
			strconv.FormatFloat(val.Bobot, 'f', -1, 64),
			ketAtribut,
		})
	}
	tableRenderer.Render()

	// Alternatif
	// Mengambil data alternatif dari file alternatif.csv
	fmt.Println("Mengambil data alternatif...")
	csvFile, err := os.Open("data/alternatif.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()

	// Mapping data csv alternatif kedalam struct alternatif
	var alternatif []Alternatif
	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		fmt.Println(err)
	}
	for _, line := range csvLines {
		tempNilaiKriteria := []float64{
			nilaiKriteriaToFloat64(line[1]),
			nilaiKriteriaToFloat64(line[2]),
			nilaiKriteriaToFloat64(line[3]),
			nilaiKriteriaToFloat64(line[4]),
			nilaiKriteriaToFloat64(line[5]),
			nilaiKriteriaToFloat64(line[6]),
		}
		temp := Alternatif{
			NamaMahasiswa: line[0],
			NilaiKriteria: tempNilaiKriteria,
		}
		alternatif = append(alternatif, temp)
	}
	// Menampilkan table alternatif
	tableRenderer = tablewriter.NewWriter(os.Stdout)
	tableRenderer.SetHeader([]string{"Nama", "C1", "C2", "C3", "C4", "C5", "C6"})
	for _, val := range alternatif {
		temp := []string{
			val.NamaMahasiswa,
		}
		for _, nilaiKriteria := range val.NilaiKriteria {
			temp = append(temp, strconv.FormatFloat(nilaiKriteria, 'f', -1, 64))
		}
		tableRenderer.Append(temp)
	}
	tableRenderer.Render()

	// Perhitungan dengan metode SAW
	fmt.Println("Proses pemeringkatan...")
	// 1. Normalisasi
	for indexK, k := range kriteria {
		if k.Atribut {
			// Kriteria benefit
			// Cari nilai maksimal kriteria
			max := alternatif[0].NilaiKriteria[indexK] // Init nilai max dengan kriteria pada alternatif ke-0
			for indexA := range alternatif {
				a := &alternatif[indexA]
				if a.NilaiKriteria[indexK] >= max {
					max = a.NilaiKriteria[indexK]
				}
			}
			// Ubah nilai kriteria menjadi nilai_kriteria_lama / max
			for indexA := range alternatif {
				a := &alternatif[indexA]
				nilaiBaru := a.NilaiKriteria[indexK] / max
				a.UbahNilaiKriteria(indexK, nilaiBaru)
			}
		} else {
			// Kriteria cost
			// Cari nilai minimal kriteria
			min := alternatif[0].NilaiKriteria[indexK] // Init nilai min dengan kriteria pada alternatif ke-0
			for indexA := range alternatif {
				a := &alternatif[indexA]
				if a.NilaiKriteria[indexK] <= min {
					min = a.NilaiKriteria[indexK]
				}
			}
			// Ubah nilai kriteria menjadi min / nilai_kriteria_lama
			for indexA := range alternatif {
				a := &alternatif[indexA]
				nilaiBaru := min / a.NilaiKriteria[indexK]
				a.UbahNilaiKriteria(indexK, nilaiBaru)
			}
		}
	}

	// 2. Perkalian Bobot
	for indexA := range alternatif {
		a := &alternatif[indexA]
		var tempTotalNilai float64
		for indexK, k := range kriteria {
			tempTotalNilai += (a.NilaiKriteria[indexK] * k.Bobot)
		}
		a.TotalNilai = tempTotalNilai
	}

	// 3. Perankingan
	// Sorting ranking tertinggi
	sort.Slice(alternatif, func(i, j int) bool {
		return alternatif[i].TotalNilai > alternatif[j].TotalNilai
	})
	// Tampilkan table
	tableRenderer = tablewriter.NewWriter(os.Stdout)
	tableRenderer.SetHeader([]string{"Ranking", "Nama", "Skor"})
	for rank, val := range alternatif {
		tableRenderer.Append([]string{
			strconv.FormatInt(int64(rank+1), 10),
			val.NamaMahasiswa,
			strconv.FormatFloat(val.TotalNilai, 'f', -1, 64),
		})
	}
	tableRenderer.Render()
}

func nilaiKriteriaToFloat64(nilai string) float64 {
	val, _ := strconv.ParseFloat(nilai, 64)
	return val
}
