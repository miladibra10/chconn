package column_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vahid-sohrabloo/chconn"
	"github.com/vahid-sohrabloo/chconn/column"
)

func TestInt16(t *testing.T) {
	t.Parallel()

	connString := os.Getenv("CHX_TEST_TCP_CONN_STRING")

	conn, err := chconn.Connect(context.Background(), connString)
	require.NoError(t, err)

	res, err := conn.Exec(context.Background(), `DROP TABLE IF EXISTS test_int16`)
	require.NoError(t, err)
	require.Nil(t, res)

	res, err = conn.Exec(context.Background(), `CREATE TABLE test_int16 (
				int16 Int16,
				int16_nullable Nullable(Int16),
				int16_array Array(Int16),
				int16_array_nullable Array(Nullable(Int16))
			) Engine=Memory`)

	require.NoError(t, err)
	require.Nil(t, res)

	col := column.NewInt16(false)

	colArrayValues := column.NewInt16(false)
	colArray := column.NewArray(colArrayValues)

	colArrayValuesNil := column.NewInt16(true)
	colArrayNil := column.NewArray(colArrayValuesNil)

	colNil := column.NewInt16(true)

	var colInsert []int16
	var colInsertArray [][]int16
	var colInsertArrayNil [][]*int16
	var colNilInsert []*int16
	for insertN := 0; insertN < 2; insertN++ {
		rows := 10
		col.Reset()
		colArrayValues.Reset()
		colArray.Reset()
		colArrayValuesNil.Reset()
		colArrayNil.Reset()
		colNil.Reset()
		for i := 0; i < rows; i++ {
			val := int16(i * -2)
			valArray := []int16{val, int16(i*-2) + 1}
			valArrayNil := []*int16{&val, nil}

			col.Append(val)
			colInsert = append(colInsert, val)

			// example insert array
			colInsertArray = append(colInsertArray, valArray)
			colArray.AppendLen(len(valArray))
			for _, v := range valArray {
				colArrayValues.Append(v)
			}

			// example insert nullable array
			colInsertArrayNil = append(colInsertArrayNil, valArrayNil)
			colArrayNil.AppendLen(len(valArrayNil))
			for _, v := range valArrayNil {
				colArrayValuesNil.AppendP(v)
			}

			// example add nullable
			if i%2 == 0 {
				colNilInsert = append(colNilInsert, &val)
				if i <= rows/2 {
					// example to add by pointer
					colNil.AppendP(&val)
				} else {
					// example to without pointer
					colNil.Append(val)
					colNil.AppendIsNil(false)
				}
			} else {
				colNilInsert = append(colNilInsert, nil)
				if i <= rows/2 {
					// example to add by pointer
					colNil.AppendP(nil)
				} else {
					// example to add without pointer
					colNil.AppendEmpty()
					colNil.AppendIsNil(true)
				}
			}
		}

		err = conn.Insert(context.Background(), `INSERT INTO
			test_int16 (int16,int16_nullable,int16_array,int16_array_nullable)
		VALUES`,
			col,
			colNil,
			colArray,
			colArrayNil,
		)

		require.NoError(t, err)
	}

	// example read all
	selectStmt, err := conn.Select(context.Background(), `SELECT
		int16,int16_nullable,int16_array,int16_array_nullable
	FROM test_int16`)
	require.NoError(t, err)
	require.True(t, conn.IsBusy())

	colRead := column.NewInt16(false)
	colNilRead := column.NewInt16(true)
	colArrayReadData := column.NewInt16(false)
	colArrayRead := column.NewArray(colArrayReadData)
	colArrayReadDataNil := column.NewInt16(true)
	colArrayReadNil := column.NewArray(colArrayReadDataNil)
	var colData []int16
	var colNilData []*int16
	var colArrayData [][]int16
	var colArrayDataNil [][]*int16

	var colArrayLens []int

	for selectStmt.Next() {
		err = selectStmt.NextColumn(colRead)
		require.NoError(t, err)
		colRead.ReadAll(&colData)

		err = selectStmt.NextColumn(colNilRead)
		require.NoError(t, err)
		colNilRead.ReadAllP(&colNilData)

		// read array
		colArrayLens = colArrayLens[:0]
		err = selectStmt.NextColumn(colArrayRead)
		require.NoError(t, err)
		colArrayRead.ReadAll(&colArrayLens)

		for _, l := range colArrayLens {
			arr := make([]int16, l)
			colArrayReadData.Fill(arr)
			colArrayData = append(colArrayData, arr)
		}

		// read nullable array
		colArrayLens = colArrayLens[:0]
		err = selectStmt.NextColumn(colArrayReadNil)
		require.NoError(t, err)
		colArrayRead.ReadAll(&colArrayLens)

		for _, l := range colArrayLens {
			arr := make([]*int16, l)
			colArrayReadDataNil.FillP(arr)
			colArrayDataNil = append(colArrayDataNil, arr)
		}
	}

	assert.Equal(t, colInsert, colData)
	assert.Equal(t, colNilInsert, colNilData)
	assert.Equal(t, colInsertArray, colArrayData)
	assert.Equal(t, colInsertArrayNil, colArrayDataNil)
	require.NoError(t, selectStmt.Err())

	selectStmt.Close()

	// example one by one
	selectStmt, err = conn.Select(context.Background(), `SELECT
		int16,int16_nullable,int16_array,int16_array_nullable
	FROM test_int16`)
	require.NoError(t, err)
	require.True(t, conn.IsBusy())

	colRead = column.NewInt16(false)
	colNilRead = column.NewInt16(true)
	colArrayReadData = column.NewInt16(false)
	colArrayRead = column.NewArray(colArrayReadData)
	colArrayReadDataNil = column.NewInt16(true)
	colArrayReadNil = column.NewArray(colArrayReadDataNil)
	colData = colData[:0]
	colNilData = colNilData[:0]
	colArrayData = colArrayData[:0]
	colArrayDataNil = colArrayDataNil[:0]

	for selectStmt.Next() {
		err = selectStmt.NextColumn(colRead)
		require.NoError(t, err)
		for colRead.Next() {
			colData = append(colData, colRead.Value())
		}

		// read nullable
		err = selectStmt.NextColumn(colNilRead)
		require.NoError(t, err)
		for colNilRead.Next() {
			colNilData = append(colNilData, colNilRead.ValueP())
		}

		// read array
		err = selectStmt.NextColumn(colArrayRead)
		require.NoError(t, err)
		for colArrayRead.Next() {
			arr := make([]int16, colArrayRead.Value())
			colArrayReadData.Fill(arr)
			colArrayData = append(colArrayData, arr)
		}

		// read nullable array
		err = selectStmt.NextColumn(colArrayReadNil)
		require.NoError(t, err)
		for colArrayReadNil.Next() {
			arr := make([]*int16, colArrayReadNil.Value())
			colArrayReadDataNil.FillP(arr)
			colArrayDataNil = append(colArrayDataNil, arr)
		}
	}

	assert.Equal(t, colInsert, colData)
	assert.Equal(t, colNilInsert, colNilData)
	assert.Equal(t, colInsertArray, colArrayData)
	assert.Equal(t, colInsertArrayNil, colArrayDataNil)
	require.NoError(t, selectStmt.Err())

	selectStmt.Close()

	conn.Close(context.Background())
}
