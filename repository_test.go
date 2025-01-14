package git

import (
	"fmt"

	"gopkg.in/src-d/go-git.v3/clients/http"
	"gopkg.in/src-d/go-git.v3/core"

	. "gopkg.in/check.v1"
)

type SuiteRepository struct {
	repos map[string]*Repository
}

var _ = Suite(&SuiteRepository{})

func (s *SuiteRepository) SetUpTest(c *C) {
	s.repos = unpackFixtures(c, tagFixtures)
}

func (s *SuiteRepository) TestNewRepository(c *C) {
	r, err := NewRepository(RepositoryFixture, nil)
	c.Assert(err, IsNil)
	c.Assert(r.Remotes["origin"].Auth, IsNil)
	c.Assert(r.URL, Equals, RepositoryFixture)
}

func (s *SuiteRepository) TestNewRepositoryWithAuth(c *C) {
	auth := &http.BasicAuth{}
	r, err := NewRepository(RepositoryFixture, auth)
	c.Assert(err, IsNil)
	c.Assert(r.Remotes["origin"].Auth, Equals, auth)
}

func (s *SuiteRepository) TestPull(c *C) {
	r, err := NewRepository(RepositoryFixture, nil)
	r.Remotes["origin"].upSrv = &MockGitUploadPackService{}

	c.Assert(err, IsNil)
	c.Assert(r.Pull("origin", "refs/heads/master"), IsNil)

	mock, ok := (r.Remotes["origin"].upSrv).(*MockGitUploadPackService)
	c.Assert(ok, Equals, true)
	err = mock.RC.Close()
	c.Assert(err, Not(IsNil), Commentf("pull leaks an open fd from the fetch"))
}

func (s *SuiteRepository) TestCommit(c *C) {
	r, err := NewRepository(RepositoryFixture, nil)
	r.Remotes["origin"].upSrv = &MockGitUploadPackService{}

	c.Assert(err, IsNil)
	c.Assert(r.Pull("origin", "refs/heads/master"), IsNil)

	hash := core.NewHash("b8e471f58bcbca63b07bda20e428190409c2db47")
	commit, err := r.Commit(hash)
	c.Assert(err, IsNil)

	c.Assert(commit.Hash.IsZero(), Equals, false)
	c.Assert(commit.Tree().Hash.IsZero(), Equals, false)
	c.Assert(commit.Author.Email, Equals, "daniel@lordran.local")
}

func (s *SuiteRepository) TestCommits(c *C) {
	r, err := NewRepository(RepositoryFixture, nil)
	r.Remotes["origin"].upSrv = &MockGitUploadPackService{}

	c.Assert(err, IsNil)
	c.Assert(r.Pull("origin", "refs/heads/master"), IsNil)

	count := 0
	commits := r.Commits()
	for {
		commit, err := commits.Next()
		if err != nil {
			break
		}

		count++
		c.Assert(commit.Hash.IsZero(), Equals, false)
		//c.Assert(commit.Tree.IsZero(), Equals, false)
	}

	c.Assert(count, Equals, 8)
}

func (s *SuiteRepository) TestTag(c *C) {
	for i, t := range tagTests {
		r, ok := s.repos[t.repo]
		c.Assert(ok, Equals, true)
		k := 0
		for hash, expected := range t.tags {
			tag, err := r.Tag(core.NewHash(hash))
			c.Assert(err, IsNil)
			testTagExpected(c, tag, expected, fmt.Sprintf("subtest %d, tag %d: ", i, k))
			k++
		}
	}
}

func (s *SuiteRepository) TestTags(c *C) {
	for i, t := range tagTests {
		r, ok := s.repos[t.repo]
		c.Assert(ok, Equals, true)
		testTagIter(c, r.Tags(), t.tags, fmt.Sprintf("subtest %d, ", i))
	}
}

func (s *SuiteRepository) TestCommitIterClosePanic(c *C) {
	r, err := NewRepository(RepositoryFixture, nil)
	r.Remotes["origin"].upSrv = &MockGitUploadPackService{}

	c.Assert(err, IsNil)
	c.Assert(r.Pull("origin", "refs/heads/master"), IsNil)

	commits := r.Commits()
	commits.Close()
}
