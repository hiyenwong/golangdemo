package xlsx

import (
	"bytes"
	"encoding/xml"
	// "strconv"
	"strings"

	. "gopkg.in/check.v1"
)

type LibSuite struct{}

var _ = Suite(&LibSuite{})

// which they are contained from the XLSX file, even when the
// worksheet files have arbitrary, non-numeric names.
func (l *LibSuite) TestReadWorkbookRelationsFromZipFileWithFunnyNames(c *C) {
	var xlsxFile *File
	var err error

	xlsxFile, err = OpenFile("./testdocs/testrels.xlsx")
	c.Assert(err, IsNil)
	bob := xlsxFile.Sheet["Bob"]
	row1 := bob.Rows[0]
	cell1 := row1.Cells[0]
	c.Assert(cell1.String(), Equals, "I am Bob")
}

// We can marshal WorkBookRels to an xml file
func (l *LibSuite) TestWorkBookRelsMarshal(c *C) {
	var rels WorkBookRels = make(WorkBookRels)
	rels["rId1"] = "worksheets/sheet.xml"
	expectedXML := `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Target="worksheets/sheet.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet"></Relationship><Relationship Id="rId2" Target="sharedStrings.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings"></Relationship><Relationship Id="rId3" Target="theme/theme1.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme"></Relationship><Relationship Id="rId4" Target="styles.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"></Relationship></Relationships>`
	xRels := rels.MakeXLSXWorkbookRels()

	output := bytes.NewBufferString(xml.Header)
	body, err := xml.Marshal(xRels)
	c.Assert(err, IsNil)
	c.Assert(body, NotNil)
	_, err = output.Write(body)
	c.Assert(err, IsNil)
	c.Assert(output.String(), Equals, expectedXML)
}

// Excel column codes are a special form of base26 that doesn't allow
// zeros, except in the least significant part of the code.  Test we
// can smoosh the numbers in a normal base26 representation (presented
// as a slice of integers) down to this form.
func (l *LibSuite) TestSmooshBase26Slice(c *C) {
	input := []int{20, 0, 1}
	expected := []int{19, 26, 1}
	c.Assert(smooshBase26Slice(input), DeepEquals, expected)
}

// formatColumnName converts slices of base26 integers to alphabetical
// column names.  Note that the least signifcant character has a
// different numeric offset (Yuck!)
func (l *LibSuite) TestFormatColumnName(c *C) {
	c.Assert(formatColumnName([]int{0}), Equals, "A")
	c.Assert(formatColumnName([]int{25}), Equals, "Z")
	c.Assert(formatColumnName([]int{1, 25}), Equals, "AZ")
	c.Assert(formatColumnName([]int{26, 25}), Equals, "ZZ")
	c.Assert(formatColumnName([]int{26, 26, 25}), Equals, "ZZZ")
}

// getLargestDenominator returns the largest power of a provided value
// that can fit within a given value.
func (l *LibSuite) TestGetLargestDenominator(c *C) {
	d, p := getLargestDenominator(0, 1, 2, 0)
	c.Assert(d, Equals, 1)
	c.Assert(p, Equals, 0)
	d, p = getLargestDenominator(1, 1, 2, 0)
	c.Assert(d, Equals, 1)
	c.Assert(p, Equals, 0)
	d, p = getLargestDenominator(2, 1, 2, 0)
	c.Assert(d, Equals, 2)
	c.Assert(p, Equals, 1)
	d, p = getLargestDenominator(4, 1, 2, 0)
	c.Assert(d, Equals, 4)
	c.Assert(p, Equals, 2)
	d, p = getLargestDenominator(8, 1, 2, 0)
	c.Assert(d, Equals, 8)
	c.Assert(p, Equals, 3)
	d, p = getLargestDenominator(9, 1, 2, 0)
	c.Assert(d, Equals, 8)
	c.Assert(p, Equals, 3)
	d, p = getLargestDenominator(15, 1, 2, 0)
	c.Assert(d, Equals, 8)
	c.Assert(p, Equals, 3)
	d, p = getLargestDenominator(16, 1, 2, 0)
	c.Assert(d, Equals, 16)
	c.Assert(p, Equals, 4)
}

func (l *LibSuite) TestLettersToNumeric(c *C) {
	cases := map[string]int{"A": 0, "G": 6, "z": 25, "AA": 26, "Az": 51,
		"BA": 52, "BZ": 77, "ZA": 26*26 + 0, "ZZ": 26*26 + 25,
		"AAA": 26*26 + 26 + 0, "AMI": 1022}
	for input, ans := range cases {
		output := lettersToNumeric(input)
		c.Assert(output, Equals, ans)
	}
}

func (l *LibSuite) TestNumericToLetters(c *C) {
	cases := map[string]int{
		"A":  0,
		"G":  6,
		"Z":  25,
		"AA": 26,
		"AZ": 51,
		"BA": 52,
		"BZ": 77, "ZA": 26 * 26, "ZB": 26*26 + 1,
		"ZZ":  26*26 + 25,
		"AAA": 26*26 + 26 + 0, "AMI": 1022}
	for ans, input := range cases {
		output := numericToLetters(input)
		c.Assert(output, Equals, ans)
	}

}

func (l *LibSuite) TestLetterOnlyMapFunction(c *C) {
	var input string = "ABC123"
	var output string = strings.Map(letterOnlyMapF, input)
	c.Assert(output, Equals, "ABC")
	input = "abc123"
	output = strings.Map(letterOnlyMapF, input)
	c.Assert(output, Equals, "ABC")
}

func (l *LibSuite) TestIntOnlyMapFunction(c *C) {
	var input string = "ABC123"
	var output string = strings.Map(intOnlyMapF, input)
	c.Assert(output, Equals, "123")
}

func (l *LibSuite) TestGetCoordsFromCellIDString(c *C) {
	var cellIDString string = "A3"
	var x, y int
	var err error
	x, y, err = getCoordsFromCellIDString(cellIDString)
	c.Assert(err, IsNil)
	c.Assert(x, Equals, 0)
	c.Assert(y, Equals, 2)
}

func (l *LibSuite) TestGetCellIDStringFromCoords(c *C) {
	c.Assert(getCellIDStringFromCoords(0, 0), Equals, "A1")
	c.Assert(getCellIDStringFromCoords(2, 2), Equals, "C3")
}

func (l *LibSuite) TestGetMaxMinFromDimensionRef(c *C) {
	var dimensionRef string = "A1:B2"
	var minx, miny, maxx, maxy int
	var err error
	minx, miny, maxx, maxy, err = getMaxMinFromDimensionRef(dimensionRef)
	c.Assert(err, IsNil)
	c.Assert(minx, Equals, 0)
	c.Assert(miny, Equals, 0)
	c.Assert(maxx, Equals, 1)
	c.Assert(maxy, Equals, 1)
}

func (l *LibSuite) TestCalculateMaxMinFromWorksheet(c *C) {
	var sheetxml = bytes.NewBufferString(`
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
           xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"
           xmlns:mx="http://schemas.microsoft.com/office/mac/excel/2008/main"
           xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006"
           xmlns:mv="urn:schemas-microsoft-com:mac:vml"
           xmlns:x14="http://schemas.microsoft.com/office/spreadsheetml/2009/9/main"
           xmlns:x14ac="http://schemas.microsoft.com/office/spreadsheetml/2009/9/ac"
           xmlns:xm="http://schemas.microsoft.com/office/excel/2006/main">
  <sheetViews>
    <sheetView workbookViewId="0"/>
  </sheetViews>
  <sheetFormatPr customHeight="1" defaultColWidth="14.43" defaultRowHeight="15.75"/>
  <sheetData>
    <row r="1">
      <c t="s" s="1" r="A1">
        <v>0</v>
      </c>
      <c t="s" s="1" r="B1">
        <v>1</v>
      </c>
    </row>
    <row r="2">
      <c t="s" s="1" r="A2">
        <v>2</v>
      </c>
      <c t="s" s="1" r="B2">
        <v>3</v>
      </c>
    </row>
  </sheetData>
  <drawing r:id="rId1"/>
</worksheet>`)
	worksheet := new(xlsxWorksheet)
	err := xml.NewDecoder(sheetxml).Decode(worksheet)
	c.Assert(err, IsNil)
	minx, miny, maxx, maxy, err := calculateMaxMinFromWorksheet(worksheet)
	c.Assert(err, IsNil)
	c.Assert(minx, Equals, 0)
	c.Assert(miny, Equals, 0)
	c.Assert(maxx, Equals, 1)
	c.Assert(maxy, Equals, 1)
}

func (l *LibSuite) TestGetRangeFromString(c *C) {
	var rangeString string
	var lower, upper int
	var err error
	rangeString = "1:3"
	lower, upper, err = getRangeFromString(rangeString)
	c.Assert(err, IsNil)
	c.Assert(lower, Equals, 1)
	c.Assert(upper, Equals, 3)
}

func (l *LibSuite) TestMakeRowFromSpan(c *C) {
	var rangeString string
	var row *Row
	var length int
	rangeString = "1:3"
	row = makeRowFromSpan(rangeString)
	length = len(row.Cells)
	c.Assert(length, Equals, 3)
	rangeString = "5:7" // Note - we ignore lower bound!
	row = makeRowFromSpan(rangeString)
	length = len(row.Cells)
	c.Assert(length, Equals, 7)
	rangeString = "1:1"
	row = makeRowFromSpan(rangeString)
	length = len(row.Cells)
	c.Assert(length, Equals, 1)
}

func (l *LibSuite) TestReadRowsFromSheet(c *C) {
	var sharedstringsXML = bytes.NewBufferString(`
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="4" uniqueCount="4">
  <si>
    <t>Foo</t>
  </si>
  <si>
    <t>Bar</t>
  </si>
  <si>
    <t xml:space="preserve">Baz </t>
  </si>
  <si>
    <t>Quuk</t>
  </si>
</sst>`)
	var sheetxml = bytes.NewBufferString(`
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
           xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <dimension ref="A1:B2"/>
  <sheetViews>
    <sheetView tabSelected="1" workbookViewId="0">
      <selection activeCell="C2" sqref="C2"/>
	  <pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/>
    </sheetView>
  </sheetViews>
  <sheetFormatPr baseColWidth="10" defaultRowHeight="15"/>
  <sheetData>
    <row r="1" spans="1:2">
      <c r="A1" t="s">
        <v>0</v>
      </c>
      <c r="B1" t="s">
        <v>1</v>
      </c>
    </row>
    <row r="2" spans="1:2">
      <c r="A2" t="s">
        <v>2</v>
      </c>
      <c r="B2" t="s">
        <v>3</v>
      </c>
    </row>
  </sheetData>
  <pageMargins left="0.7" right="0.7"
               top="0.78740157499999996"
               bottom="0.78740157499999996"
               header="0.3"
               footer="0.3"/>
</worksheet>`)
	worksheet := new(xlsxWorksheet)
	err := xml.NewDecoder(sheetxml).Decode(worksheet)
	c.Assert(err, IsNil)
	sst := new(xlsxSST)
	err = xml.NewDecoder(sharedstringsXML).Decode(sst)
	c.Assert(err, IsNil)
	file := new(File)
	file.referenceTable = MakeSharedStringRefTable(sst)
	rows, cols, maxCols, maxRows := readRowsFromSheet(worksheet, file)
	c.Assert(maxRows, Equals, 2)
	c.Assert(maxCols, Equals, 2)
	row := rows[0]
	c.Assert(len(row.Cells), Equals, 2)
	cell1 := row.Cells[0]
	c.Assert(cell1.Value, Equals, "Foo")
	cell2 := row.Cells[1]
	c.Assert(cell2.Value, Equals, "Bar")
	col := cols[0]
	c.Assert(col.Min, Equals, 0)
	c.Assert(col.Max, Equals, 0)
	c.Assert(col.Hidden, Equals, false)
	c.Assert(len(worksheet.SheetViews.SheetView), Equals, 1)
	sheetView := worksheet.SheetViews.SheetView[0]
	c.Assert(sheetView.Pane, NotNil)
	pane := sheetView.Pane
	c.Assert(pane.XSplit, Equals, 0)
	c.Assert(pane.YSplit, Equals, 1)
}

func (l *LibSuite) TestReadRowsFromSheetWithLeadingEmptyRows(c *C) {
	var sharedstringsXML = bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="2" uniqueCount="2"><si><t>ABC</t></si><si><t>DEF</t></si></sst>`)
	var sheetxml = bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" mc:Ignorable="x14ac" xmlns:x14ac="http://schemas.microsoft.com/office/spreadsheetml/2009/9/ac">
  <dimension ref="A4:A5"/>
  <sheetViews>
    <sheetView tabSelected="1" workbookViewId="0">
      <selection activeCell="A2" sqref="A2"/>
    </sheetView>
  </sheetViews>
  <sheetFormatPr baseColWidth="10" defaultRowHeight="15" x14ac:dyDescent="0"/>
  <sheetData>
    <row r="4" spans="1:1">
      <c r="A4" t="s">
        <v>0</v>
      </c>
    </row>
    <row r="5" spans="1:1">
      <c r="A5" t="s">
        <v>1</v>
      </c>
    </row>
  </sheetData>
  <pageMargins left="0.75" right="0.75" top="1" bottom="1" header="0.5" footer="0.5"/>
  <pageSetup paperSize="9" orientation="portrait" horizontalDpi="4294967292" verticalDpi="4294967292"/>
  <extLst>
    <ext uri="{64002731-A6B0-56B0-2670-7721B7C09600}" xmlns:mx="http://schemas.microsoft.com/office/mac/excel/2008/main">
      <mx:PLV Mode="0" OnePage="0" WScale="0"/>
    </ext>
  </extLst>
</worksheet>
`)
	worksheet := new(xlsxWorksheet)
	err := xml.NewDecoder(sheetxml).Decode(worksheet)
	c.Assert(err, IsNil)
	sst := new(xlsxSST)
	err = xml.NewDecoder(sharedstringsXML).Decode(sst)
	c.Assert(err, IsNil)

	file := new(File)
	file.referenceTable = MakeSharedStringRefTable(sst)
	rows, _, maxCols, maxRows := readRowsFromSheet(worksheet, file)
	c.Assert(maxRows, Equals, 5)
	c.Assert(maxCols, Equals, 1)

	c.Assert(len(rows[0].Cells), Equals, 0)
	c.Assert(len(rows[1].Cells), Equals, 0)
	c.Assert(len(rows[2].Cells), Equals, 0)
	c.Assert(len(rows[3].Cells), Equals, 1)
	c.Assert(rows[3].Cells[0].String(), Equals, "ABC")
	c.Assert(len(rows[4].Cells), Equals, 1)
	c.Assert(rows[4].Cells[0].String(), Equals, "DEF")
}

func (l *LibSuite) TestReadRowsFromSheetWithLeadingEmptyCols(c *C) {
	var sharedstringsXML = bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="2" uniqueCount="2"><si><t>ABC</t></si><si><t>DEF</t></si></sst>`)
	var sheetxml = bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" mc:Ignorable="x14ac" xmlns:x14ac="http://schemas.microsoft.com/office/spreadsheetml/2009/9/ac">
  <dimension ref="C1:D2"/>
  <sheetViews>
    <sheetView tabSelected="1" workbookViewId="0">
      <selection activeCell="A2" sqref="A2"/>
    </sheetView>
  </sheetViews>
  <sheetFormatPr baseColWidth="10" defaultRowHeight="15" x14ac:dyDescent="0"/>
  <cols>
  	<col min="3" max="3" width="17" customWidth="1"/>
  	<col min="4" max="4" width="18" customWidth="1"/>
  </cols>
  <sheetData>
    <row r="1" spans="3:4">
      <c r="C1" t="s"><v>0</v></c>
      <c r="D1" t="s"><v>1</v></c>
    </row>
    <row r="2" spans="3:4">
      <c r="C2" t="s"><v>0</v></c>
      <c r="D2" t="s"><v>1</v></c>
    </row>
  </sheetData>
  <pageMargins left="0.75" right="0.75" top="1" bottom="1" header="0.5" footer="0.5"/>
  <pageSetup paperSize="9" orientation="portrait" horizontalDpi="4294967292" verticalDpi="4294967292"/>
  <extLst>
    <ext uri="{64002731-A6B0-56B0-2670-7721B7C09600}" xmlns:mx="http://schemas.microsoft.com/office/mac/excel/2008/main">
      <mx:PLV Mode="0" OnePage="0" WScale="0"/>
    </ext>
  </extLst>
</worksheet>
`)
	worksheet := new(xlsxWorksheet)
	err := xml.NewDecoder(sheetxml).Decode(worksheet)
	c.Assert(err, IsNil)
	sst := new(xlsxSST)
	err = xml.NewDecoder(sharedstringsXML).Decode(sst)
	c.Assert(err, IsNil)

	file := new(File)
	file.referenceTable = MakeSharedStringRefTable(sst)
	rows, cols, maxCols, maxRows := readRowsFromSheet(worksheet, file)
	c.Assert(maxRows, Equals, 2)
	c.Assert(maxCols, Equals, 4)

	c.Assert(len(rows[0].Cells), Equals, 4)
	c.Assert(rows[0].Cells[0].String(), Equals, "")
	c.Assert(rows[0].Cells[1].String(), Equals, "")
	c.Assert(rows[0].Cells[2].String(), Equals, "ABC")
	c.Assert(rows[0].Cells[3].String(), Equals, "DEF")
	c.Assert(len(rows[1].Cells), Equals, 4)
	c.Assert(rows[1].Cells[0].String(), Equals, "")
	c.Assert(rows[1].Cells[1].String(), Equals, "")
	c.Assert(rows[1].Cells[2].String(), Equals, "ABC")
	c.Assert(rows[1].Cells[3].String(), Equals, "DEF")

	c.Assert(len(cols), Equals, 4)
	c.Assert(cols[0].Width, Equals, 0.0)
	c.Assert(cols[1].Width, Equals, 0.0)
	c.Assert(cols[2].Width, Equals, 17.0)
	c.Assert(cols[3].Width, Equals, 18.0)
}

func (l *LibSuite) TestReadRowsFromSheetWithEmptyCells(c *C) {
	var sharedstringsXML = bytes.NewBufferString(`
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="8" uniqueCount="5">
  <si>
    <t>Bob</t>
  </si>
  <si>
    <t>Alice</t>
  </si>
  <si>
    <t>Sue</t>
  </si>
  <si>
    <t>Yes</t>
  </si>
  <si>
    <t>No</t>
  </si>
</sst>
`)
	var sheetxml = bytes.NewBufferString(`
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><dimension ref="A1:C3"/><sheetViews><sheetView tabSelected="1" workbookViewId="0"><selection activeCell="D3" sqref="D3"/></sheetView></sheetViews><sheetFormatPr baseColWidth="10" defaultRowHeight="15"/>
<sheetData>
  <row r="1" spans="1:3">
    <c r="A1" t="s">
      <v>
        0
      </v>
    </c>
    <c r="B1" t="s">
      <v>
        1
      </v>
    </c>
    <c r="C1" t="s">
      <v>
        2
      </v>
    </c>
  </row>
  <row r="2" spans="1:3">
    <c r="A2" t="s">
      <v>
        3
      </v>
    </c>
    <c r="B2" t="s">
      <v>
        4
      </v>
    </c>
    <c r="C2" t="s">
      <v>
        3
      </v>
    </c>
  </row>
  <row r="3" spans="1:3">
    <c r="A3" t="s">
      <v>
        4
      </v>
    </c>
    <c r="C3" t="s">
      <v>
        3
      </v>
    </c>
  </row>
</sheetData>
<pageMargins left="0.7" right="0.7" top="0.78740157499999996" bottom="0.78740157499999996" header="0.3" footer="0.3"/>
</worksheet>

`)
	worksheet := new(xlsxWorksheet)
	err := xml.NewDecoder(sheetxml).Decode(worksheet)
	c.Assert(err, IsNil)
	sst := new(xlsxSST)
	err = xml.NewDecoder(sharedstringsXML).Decode(sst)
	c.Assert(err, IsNil)
	file := new(File)
	file.referenceTable = MakeSharedStringRefTable(sst)
	rows, cols, maxCols, maxRows := readRowsFromSheet(worksheet, file)
	c.Assert(maxRows, Equals, 3)
	c.Assert(maxCols, Equals, 3)

	row := rows[2]
	c.Assert(len(row.Cells), Equals, 3)

	cell1 := row.Cells[0]
	c.Assert(cell1.Value, Equals, "No")

	cell2 := row.Cells[1]
	c.Assert(cell2.Value, Equals, "")

	cell3 := row.Cells[2]
	c.Assert(cell3.Value, Equals, "Yes")

	col := cols[0]
	c.Assert(col.Min, Equals, 0)
	c.Assert(col.Max, Equals, 0)
	c.Assert(col.Hidden, Equals, false)
}

func (l *LibSuite) TestReadRowsFromSheetWithTrailingEmptyCells(c *C) {
	var row *Row
	var cell1, cell2, cell3, cell4 *Cell
	var sharedstringsXML = bytes.NewBufferString(`
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="4" uniqueCount="4"><si><t>A</t></si><si><t>B</t></si><si><t>C</t></si><si><t>D</t></si></sst>`)
	var sheetxml = bytes.NewBufferString(`
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><dimension ref="A1:D8"/><sheetViews><sheetView tabSelected="1" workbookViewId="0"><selection activeCell="A7" sqref="A7"/></sheetView></sheetViews><sheetFormatPr baseColWidth="10" defaultRowHeight="15"/><sheetData><row r="1" spans="1:4"><c r="A1" t="s"><v>0</v></c><c r="B1" t="s"><v>1</v></c><c r="C1" t="s"><v>2</v></c><c r="D1" t="s"><v>3</v></c></row><row r="2" spans="1:4"><c r="A2"><v>1</v></c></row><row r="3" spans="1:4"><c r="B3"><v>1</v></c></row><row r="4" spans="1:4"><c r="C4"><v>1</v></c></row><row r="5" spans="1:4"><c r="D5"><v>1</v></c></row><row r="6" spans="1:4"><c r="C6"><v>1</v></c></row><row r="7" spans="1:4"><c r="B7"><v>1</v></c></row><row r="8" spans="1:4"><c r="A8"><v>1</v></c></row></sheetData><pageMargins left="0.7" right="0.7" top="0.78740157499999996" bottom="0.78740157499999996" header="0.3" footer="0.3"/></worksheet>
`)
	worksheet := new(xlsxWorksheet)
	err := xml.NewDecoder(sheetxml).Decode(worksheet)
	c.Assert(err, IsNil)

	sst := new(xlsxSST)
	err = xml.NewDecoder(sharedstringsXML).Decode(sst)
	c.Assert(err, IsNil)

	file := new(File)
	file.referenceTable = MakeSharedStringRefTable(sst)
	rows, _, maxCol, maxRow := readRowsFromSheet(worksheet, file)
	c.Assert(maxCol, Equals, 4)
	c.Assert(maxRow, Equals, 8)

	row = rows[0]
	c.Assert(len(row.Cells), Equals, 4)

	cell1 = row.Cells[0]
	c.Assert(cell1.Value, Equals, "A")

	cell2 = row.Cells[1]
	c.Assert(cell2.Value, Equals, "B")

	cell3 = row.Cells[2]
	c.Assert(cell3.Value, Equals, "C")

	cell4 = row.Cells[3]
	c.Assert(cell4.Value, Equals, "D")

	row = rows[1]
	c.Assert(len(row.Cells), Equals, 4)

	cell1 = row.Cells[0]
	c.Assert(cell1.Value, Equals, "1")

	cell2 = row.Cells[1]
	c.Assert(cell2.Value, Equals, "")

	cell3 = row.Cells[2]
	c.Assert(cell3.Value, Equals, "")

	cell4 = row.Cells[3]
	c.Assert(cell4.Value, Equals, "")
}

func (l *LibSuite) TestReadRowsFromSheetWithMultipleSpans(c *C) {
	var sharedstringsXML = bytes.NewBufferString(`
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="4" uniqueCount="4">
  <si>
    <t>Foo</t>
  </si>
  <si>
    <t>Bar</t>
  </si>
  <si>
    <t xml:space="preserve">Baz </t>
  </si>
  <si>
    <t>Quuk</t>
  </si>
</sst>`)
	var sheetxml = bytes.NewBufferString(`
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
           xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <dimension ref="A1:D2"/>
  <sheetViews>
    <sheetView tabSelected="1" workbookViewId="0">
      <selection activeCell="C2" sqref="C2"/>
    </sheetView>
  </sheetViews>
  <sheetFormatPr baseColWidth="10" defaultRowHeight="15"/>
  <sheetData>
    <row r="1" spans="1:2 3:4">
      <c r="A1" t="s">
        <v>0</v>
      </c>
      <c r="B1" t="s">
        <v>1</v>
      </c>
      <c r="C1" t="s">
        <v>0</v>
      </c>
      <c r="D1" t="s">
        <v>1</v>
      </c>
    </row>
    <row r="2" spans="1:2 3:4">
      <c r="A2" t="s">
        <v>2</v>
      </c>
      <c r="B2" t="s">
        <v>3</v>
      </c>
      <c r="C2" t="s">
        <v>2</v>
      </c>
      <c r="D2" t="s">
        <v>3</v>
      </c>
    </row>
  </sheetData>
  <pageMargins left="0.7" right="0.7"
               top="0.78740157499999996"
               bottom="0.78740157499999996"
               header="0.3"
               footer="0.3"/>
</worksheet>`)
	worksheet := new(xlsxWorksheet)
	err := xml.NewDecoder(sheetxml).Decode(worksheet)
	c.Assert(err, IsNil)
	sst := new(xlsxSST)
	err = xml.NewDecoder(sharedstringsXML).Decode(sst)
	c.Assert(err, IsNil)
	file := new(File)
	file.referenceTable = MakeSharedStringRefTable(sst)
	rows, _, maxCols, maxRows := readRowsFromSheet(worksheet, file)
	c.Assert(maxRows, Equals, 2)
	c.Assert(maxCols, Equals, 4)
	row := rows[0]
	c.Assert(len(row.Cells), Equals, 4)
	cell1 := row.Cells[0]
	c.Assert(cell1.Value, Equals, "Foo")
	cell2 := row.Cells[1]
	c.Assert(cell2.Value, Equals, "Bar")
	cell3 := row.Cells[2]
	c.Assert(cell3.Value, Equals, "Foo")
	cell4 := row.Cells[3]
	c.Assert(cell4.Value, Equals, "Bar")

}

func (l *LibSuite) TestReadRowsFromSheetWithMultipleTypes(c *C) {
	var sharedstringsXML = bytes.NewBufferString(`
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="4" uniqueCount="4">
  <si>
    <t>Hello World</t>
  </si>
</sst>`)
	var sheetxml = bytes.NewBufferString(`
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
           xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <dimension ref="A1:F1"/>
  <sheetViews>
    <sheetView tabSelected="1" workbookViewId="0">
      <selection activeCell="C1" sqref="C1"/>
    </sheetView>
  </sheetViews>
  <sheetFormatPr baseColWidth="10" defaultRowHeight="15"/>
  <sheetData>
    <row r="1" spans="1:6">
      <c r="A1" t="s">
        <v>0</v>
      </c>
      <c r="B1">
        <v>12345</v>
      </c>
      <c r="C1">
        <v>1.024</v>
      </c>
      <c r="D1" t="b">
        <v>1</v>
      </c>
      <c r="E1">
      	<f>10+20</f>
        <v>30</v>
      </c>
      <c r="F1" t="e">
      	<f>10/0</f>
        <v>#DIV/0!</v>
      </c>
    </row>
  </sheetData>
  <pageMargins left="0.7" right="0.7"
               top="0.78740157499999996"
               bottom="0.78740157499999996"
               header="0.3"
               footer="0.3"/>
</worksheet>`)
	worksheet := new(xlsxWorksheet)
	err := xml.NewDecoder(sheetxml).Decode(worksheet)
	c.Assert(err, IsNil)
	sst := new(xlsxSST)
	err = xml.NewDecoder(sharedstringsXML).Decode(sst)
	c.Assert(err, IsNil)
	file := new(File)
	file.referenceTable = MakeSharedStringRefTable(sst)
	rows, _, maxCols, maxRows := readRowsFromSheet(worksheet, file)
	c.Assert(maxRows, Equals, 1)
	c.Assert(maxCols, Equals, 6)
	row := rows[0]
	c.Assert(len(row.Cells), Equals, 6)

	cell1 := row.Cells[0]
	c.Assert(cell1.Type(), Equals, CellTypeString)
	c.Assert(cell1.String(), Equals, "Hello World")

	cell2 := row.Cells[1]
	c.Assert(cell2.Type(), Equals, CellTypeNumeric)
	intValue, _ := cell2.Int()
	c.Assert(intValue, Equals, 12345)

	cell3 := row.Cells[2]
	c.Assert(cell3.Type(), Equals, CellTypeNumeric)
	float, _ := cell3.Float()
	c.Assert(float, Equals, 1.024)

	cell4 := row.Cells[3]
	c.Assert(cell4.Type(), Equals, CellTypeBool)
	c.Assert(cell4.Bool(), Equals, true)

	cell5 := row.Cells[4]
	c.Assert(cell5.Type(), Equals, CellTypeFormula)
	c.Assert(cell5.Formula(), Equals, "10+20")
	c.Assert(cell5.Value, Equals, "30")

	cell6 := row.Cells[5]
	c.Assert(cell6.Type(), Equals, CellTypeError)
	c.Assert(cell6.Formula(), Equals, "10/0")
	c.Assert(cell6.Value, Equals, "#DIV/0!")
}

// When converting the xlsxRow to a Row we create a as many cells as we find.
func (l *LibSuite) TestReadRowFromRaw(c *C) {
	var rawRow xlsxRow
	var cell xlsxC
	var row *Row

	rawRow = xlsxRow{}
	cell = xlsxC{R: "A1"}
	cell = xlsxC{R: "A2"}
	rawRow.C = append(rawRow.C, cell)
	row = makeRowFromRaw(rawRow)
	c.Assert(row, NotNil)
	c.Assert(row.Cells, HasLen, 1)
}

// When a cell claims it is at a position greater than its ordinal
// position in the file we make up the missing cells.
func (l *LibSuite) TestReadRowFromRawWithMissingCells(c *C) {
	var rawRow xlsxRow
	var cell xlsxC
	var row *Row

	rawRow = xlsxRow{}
	cell = xlsxC{R: "A1"}
	rawRow.C = append(rawRow.C, cell)
	cell = xlsxC{R: "E1"}
	rawRow.C = append(rawRow.C, cell)
	row = makeRowFromRaw(rawRow)
	c.Assert(row, NotNil)
	c.Assert(row.Cells, HasLen, 5)
}

// We can cope with missing coordinate references
func (l *LibSuite) TestReadRowFromRawWithPartialCoordinates(c *C) {
	var rawRow xlsxRow
	var cell xlsxC
	var row *Row

	rawRow = xlsxRow{}
	cell = xlsxC{R: "A1"}
	rawRow.C = append(rawRow.C, cell)
	cell = xlsxC{}
	rawRow.C = append(rawRow.C, cell)
	cell = xlsxC{R: "Z:1"}
	rawRow.C = append(rawRow.C, cell)
	cell = xlsxC{}
	rawRow.C = append(rawRow.C, cell)
	row = makeRowFromRaw(rawRow)
	c.Assert(row, NotNil)
	c.Assert(row.Cells, HasLen, 27)
}
