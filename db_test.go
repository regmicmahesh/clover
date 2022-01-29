package clover

import (
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func runCloverTest(t *testing.T, dir string, test func(t *testing.T, db *DB)) {
	if dir == "" {
		var err error
		dir, err = ioutil.TempDir("", "clover-test")
		require.NoError(t, err)
		defer os.RemoveAll(dir)
	}
	db, err := Open(dir)
	require.NoError(t, err)

	test(t, db)
}

func TestCreateCollection(t *testing.T) {
	runCloverTest(t, "", func(t *testing.T, db *DB) {
		_, err := db.CreateCollection("myCollection")
		require.NoError(t, err)
		require.True(t, db.HasCollection("myCollection"))
	})
}

func TestInsertOne(t *testing.T) {
	runCloverTest(t, "", func(t *testing.T, db *DB) {
		_, err := db.CreateCollection("myCollection")
		require.NoError(t, err)

		doc := NewDocument()
		doc.Set("hello", "clover")

		docId, err := db.InsertOne("myCollection", doc)
		require.NoError(t, err)
		require.NotEmpty(t, docId)
	})
}

func TestInsert(t *testing.T) {
	runCloverTest(t, "", func(t *testing.T, db *DB) {
		_, err := db.CreateCollection("myCollection")
		require.NoError(t, err)

		doc := NewDocument()
		doc.Set("hello", "clover")

		require.NoError(t, db.Insert("myCollection", doc))
	})
}

func TestInsertAndGet(t *testing.T) {
	runCloverTest(t, "", func(t *testing.T, db *DB) {
		c, err := db.CreateCollection("myCollection")
		require.NoError(t, err)

		nInserts := 100
		docs := make([]*Document, 0, nInserts)
		for i := 0; i < nInserts; i++ {
			doc := NewDocument()
			doc.Set("myField", i)
			docs = append(docs, doc)
		}

		require.NoError(t, db.Insert("myCollection", docs...))
		require.Equal(t, nInserts, c.Count())

		n := c.Matches(func(doc *Document) bool {
			require.True(t, doc.Has("myField"))

			v, _ := doc.Get("myField").(float64)
			return int(v)%2 == 0
		}).Count()

		require.Equal(t, nInserts/2, n)
	})
}

func copyCollection(t *testing.T, db *DB, src, dst string) error {
	if _, err := db.CreateCollection(dst); err != nil {
		return err
	}
	srcDocs := db.Query(src).FindAll()
	return db.Insert(dst, srcDocs...)
}

func TestInsertAndDelete(t *testing.T) {
	runCloverTest(t, "test-db", func(t *testing.T, db *DB) {
		err := copyCollection(t, db, "todos", "todos-temp")
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.DropCollection("todos-temp"), err)
		}()

		criteria := Row("completed").Eq(true)

		tempTodos := db.Query("todos-temp")
		require.Equal(t, tempTodos.Count(), db.Query("todos").Count())

		err = tempTodos.Where(criteria).Delete()
		require.NoError(t, err)

		// since collection is immutable, we don't see changes in old reference
		tempTodos = db.Query("todos-temp")
		require.Equal(t, tempTodos.Count(), tempTodos.Where(criteria.Not()).Count())
	})
}

func TestOpenExisting(t *testing.T) {
	runCloverTest(t, "test-db", func(t *testing.T, db *DB) {
		require.True(t, db.HasCollection("todos"))
		require.NotNil(t, db.Query("todos"))

		rows := db.Query("todos").Count()
		require.Equal(t, rows, 200)
	})
}

func TestExistsCriteria(t *testing.T) {
	runCloverTest(t, "test-db", func(t *testing.T, db *DB) {
		require.True(t, db.HasCollection("todos"))
		require.NotNil(t, db.Query("todos"))

		docs := db.Query("todos").Where(Row("completed_date").Exists()).FindAll()
		require.Equal(t, len(docs), 1)
	})
}

func TestEqCriteria(t *testing.T) {
	runCloverTest(t, "test-db", func(t *testing.T, db *DB) {
		require.True(t, db.HasCollection("todos"))
		require.NotNil(t, db.Query("todos"))

		docs := db.Query("todos").Where(Row("completed").Eq(true)).FindAll()
		require.Greater(t, len(docs), 0)

		for _, doc := range docs {
			require.NotNil(t, doc.Get("completed"))
			require.Equal(t, doc.Get("completed"), true)
		}
	})
}

func TestNeqCriteria(t *testing.T) {
	runCloverTest(t, "test-db", func(t *testing.T, db *DB) {
		require.True(t, db.HasCollection("todos"))
		require.NotNil(t, db.Query("todos"))

		docs := db.Query("todos").Where(Row("userId").Neq(7)).FindAll()
		require.Greater(t, len(docs), 0)

		for _, doc := range docs {
			require.NotNil(t, doc.Get("userId"))
			require.NotEqual(t, doc.Get("userId"), float64(7))
		}
	})
}

func TestGtCriteria(t *testing.T) {
	runCloverTest(t, "test-db", func(t *testing.T, db *DB) {
		require.True(t, db.HasCollection("todos"))
		require.NotNil(t, db.Query("todos"))

		docs := db.Query("todos").Where(Row("userId").Gt(4)).FindAll()
		require.Greater(t, len(docs), 0)

		for _, doc := range docs {
			require.NotNil(t, doc.Get("userId"))
			require.Greater(t, doc.Get("userId"), float64(4))
		}
	})
}

func TestGtEqCriteria(t *testing.T) {
	runCloverTest(t, "test-db", func(t *testing.T, db *DB) {
		require.True(t, db.HasCollection("todos"))
		require.NotNil(t, db.Query("todos"))

		docs := db.Query("todos").Where(Row("userId").GtEq(4)).FindAll()
		require.Greater(t, len(docs), 0)

		for _, doc := range docs {
			require.NotNil(t, doc.Get("userId"))
			require.GreaterOrEqual(t, doc.Get("userId"), float64(4))
		}
	})
}

func TestLtCriteria(t *testing.T) {
	runCloverTest(t, "test-db", func(t *testing.T, db *DB) {
		require.True(t, db.HasCollection("todos"))
		require.NotNil(t, db.Query("todos"))

		docs := db.Query("todos").Where(Row("userId").Lt(4)).FindAll()
		require.Greater(t, len(docs), 0)
		for _, doc := range docs {
			require.NotNil(t, doc.Get("userId"))
			require.Less(t, doc.Get("userId"), float64(4))
		}
	})
}

func TestLtEqCriteria(t *testing.T) {
	runCloverTest(t, "test-db", func(t *testing.T, db *DB) {
		require.True(t, db.HasCollection("todos"))
		require.NotNil(t, db.Query("todos"))

		docs := db.Query("todos").Where(Row("userId").LtEq(4)).FindAll()
		require.Greater(t, len(docs), 0)

		for _, doc := range docs {
			require.NotNil(t, doc.Get("userId"))
			require.LessOrEqual(t, doc.Get("userId"), float64(4))
		}
	})
}

func TestInCriteria(t *testing.T) {
	runCloverTest(t, "test-db", func(t *testing.T, db *DB) {
		require.True(t, db.HasCollection("todos"))
		require.NotNil(t, db.Query("todos"))

		docs := db.Query("todos").Where(Row("userId").In(5, 8)).FindAll()

		require.Greater(t, len(docs), 0)

		for _, doc := range docs {
			userId := doc.Get("userId")
			require.NotNil(t, userId)

			if userId != float64(5) && userId != float64(8) {
				require.Fail(t, "userId is not in the correct range")
			}
		}
	})
}

func TestAndCriteria(t *testing.T) {
	runCloverTest(t, "test-db", func(t *testing.T, db *DB) {
		require.True(t, db.HasCollection("todos"))
		require.NotNil(t, db.Query("todos"))

		criteria := Row("completed").Eq(true).And(Row("userId").Gt(2))
		docs := db.Query("todos").Where(criteria).FindAll()

		require.Greater(t, len(docs), 0)
		for _, doc := range docs {
			require.NotNil(t, doc.Get("completed"))
			require.NotNil(t, doc.Get("userId"))
			require.Equal(t, doc.Get("completed"), true)
			require.Greater(t, doc.Get("userId"), float64(2))
		}
	})
}

func genRandomFieldName() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	size := rand.Intn(100) + 1

	fName := ""
	for i := 0; i < size; i++ {
		fName += "." + string(letters[rand.Intn(len(letters))])
	}
	return fName
}

func TestDocument(t *testing.T) {
	doc := NewDocument()

	nTests := 1000
	for i := 0; i < nTests; i++ {
		fieldName := genRandomFieldName()
		doc.Set(fieldName, i)
		require.True(t, doc.Has(fieldName))
		require.Equal(t, doc.Get(fieldName), float64(i))
	}
}
