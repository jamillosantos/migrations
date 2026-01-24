//go:generate go run go.uber.org/mock/mockgen -package pgx -destination fs_mock_test.go io/fs DirEntry,ReadDirFS,File

package pgx

import (
	"context"
	"errors"
	"io"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Source", func() {
	var (
		ctrl *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("parseSQLFile", func() {
		It("should match a file with type", func() {
			dirEntry := NewMockDirEntry(ctrl)

			dirEntry.EXPECT().IsDir().Return(false)
			wantID := "123"
			wantDescription := "description with new data"
			wantType := "down"
			givenName := "123_description_with_new_data.down.sql"
			dirEntry.EXPECT().Name().Return(givenName)

			gotID, gotDescription, gotType := parseSQLFile(dirEntry)
			Expect(gotID).To(Equal(wantID))
			Expect(gotDescription).To(Equal(wantDescription))
			Expect(gotType).To(Equal(wantType))
		})

		It("should match a file without type", func() {
			dirEntry := NewMockDirEntry(ctrl)

			dirEntry.EXPECT().IsDir().Return(false)
			wantID := "123"
			wantDescription := "description with new data"
			wantType := ""
			givenName := "123_description_with_new_data.sql"
			dirEntry.EXPECT().Name().Return(givenName)

			gotID, gotDescription, gotType := parseSQLFile(dirEntry)
			Expect(gotID).To(Equal(wantID))
			Expect(gotDescription).To(Equal(wantDescription))
			Expect(gotType).To(Equal(wantType))
		})

		It("should not match when the entry is a directory", func() {
			dirEntry := NewMockDirEntry(ctrl)

			dirEntry.EXPECT().IsDir().Return(true)

			gotID, _, _ := parseSQLFile(dirEntry)
			Expect(gotID).To(BeEmpty())
		})

		It("should not match file name", func() {
			dirEntry := NewMockDirEntry(ctrl)

			dirEntry.EXPECT().IsDir().Return(false)
			dirEntry.EXPECT().Name().Return("crazy_file_name.txt")

			gotID, _, _ := parseSQLFile(dirEntry)
			Expect(gotID).To(BeEmpty())
		})
	})

	Describe("loadMigrationFile", func() {
		It("should load the migration content", func() {
			wantFile := "random file"
			fsMock := NewMockReadDirFS(ctrl)
			f := NewMockFile(ctrl)
			wantContent := "migration content"

			fsMock.EXPECT().Open(wantFile).Return(f, nil)
			f.EXPECT().Read(gomock.Any()).DoAndReturn(func(d []byte) (int, error) {
				copy(d, wantContent)
				return len(wantContent), io.EOF
			})
			f.EXPECT().Close().Return(nil)

			gotContent, err := loadMigrationFile(fsMock, wantFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(gotContent).To(Equal(wantContent))
		})

		It("should return empty when migration name is empty", func() {
			fsMock := NewMockReadDirFS(ctrl)

			content, err := loadMigrationFile(fsMock, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(BeEmpty())
		})

		It("should fail when opening migration fails", func() {
			fsMock := NewMockReadDirFS(ctrl)
			wantErr := errors.New("random error")

			fsMock.EXPECT().Open(gomock.Any()).Return(nil, wantErr)

			_, err := loadMigrationFile(fsMock, "random file")
			Expect(err).To(MatchError(wantErr))
		})

		It("should fail reading the migration content", func() {
			fsMock := NewMockReadDirFS(ctrl)
			f := NewMockFile(ctrl)
			wantErr := errors.New("random error")

			fsMock.EXPECT().Open(gomock.Any()).Return(f, nil)
			f.EXPECT().Read(gomock.Any()).Return(0, wantErr)
			f.EXPECT().Close().Return(nil)

			_, err := loadMigrationFile(fsMock, "random file")
			Expect(err).To(MatchError(wantErr))
		})
	})

	Describe("source.Load", func() {
		It("should load migrations from fs", func() {
			fsMock := NewMockReadDirFS(ctrl)
			entry1 := NewMockDirEntry(ctrl)
			entry2 := NewMockDirEntry(ctrl)

			folder := "migrations"

			entry1.EXPECT().IsDir().Return(false).AnyTimes()
			entry1.EXPECT().Name().Return("1_create_table.up.sql").AnyTimes()

			entry2.EXPECT().IsDir().Return(false).AnyTimes()
			entry2.EXPECT().Name().Return("1_create_table.down.sql").AnyTimes()

			fsMock.EXPECT().ReadDir(folder).Return([]fs.DirEntry{entry1, entry2}, nil)

			// Mock file opening for content loading
			f1 := NewMockFile(ctrl)
			fsMock.EXPECT().Open("migrations/1_create_table.up.sql").Return(f1, nil)
			f1.EXPECT().Read(gomock.Any()).DoAndReturn(func(d []byte) (int, error) {
				copy(d, "CREATE TABLE test (id int);")
				return len("CREATE TABLE test (id int);"), io.EOF
			})
			f1.EXPECT().Close().Return(nil)

			f2 := NewMockFile(ctrl)
			fsMock.EXPECT().Open("migrations/1_create_table.down.sql").Return(f2, nil)
			f2.EXPECT().Read(gomock.Any()).DoAndReturn(func(d []byte) (int, error) {
				copy(d, "DROP TABLE test;")
				return len("DROP TABLE test;"), io.EOF
			})
			f2.EXPECT().Close().Return(nil)

			s, err := SourceFromFS(nil, fsMock, folder)
			Expect(err).ToNot(HaveOccurred())

			repo, err := s.Load(context.Background())
			Expect(err).ToNot(HaveOccurred())

			m, err := repo.ByID("1")
			Expect(err).ToNot(HaveOccurred())
			Expect(m.ID()).To(Equal("1"))
			Expect(m.Description()).To(Equal("create table"))
			Expect(m.CanUndo()).To(BeTrue())

			mPgx := m.(*migrationPgx)
			Expect(mPgx.doFileContent).To(Equal("CREATE TABLE test (id int);"))
			Expect(mPgx.undoFileContent).To(Equal("DROP TABLE test;"))
		})
	})
})
