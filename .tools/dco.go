package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	flCommitRange = flag.String("range", "", "use this commit range instead")

	validDCO = regexp.MustCompile(`^Signed-off-by: ([^<]+) <([^<>@]+@[^<>]+)>$`)
)

var DefaultRules = []ValidateRule{
	func(c CommitEntry) error {
		hasValid := false
		for _, line := range strings.Split(c.Body, "\n") {
			if validDCO.MatchString(line) {
				hasValid = true
			}
		}
		if !hasValid {
			return fmt.Errorf("%q does not have a valid DCO", c.Commit)
		}
		return nil
	},
	// TODO add something for the cleanliness of the c.Subject
}

func main() {
	flag.Parse()

	/*
		if len(os.Getenv("TRAVIS")) > 0 {
			fmt.Println(os.Getenv("TRAVIS_COMMIT_RANGE"))
			if len(os.Getenv("TRAVIS_PULL_REQUEST")) > 0 {
				fmt.Println(os.Getenv("TRAVIS_BRANCH"))
			}
		}
	*/

	invalids := []InvalidCommit{}
	if *flCommitRange != "" {
		c, err := GitCommits(*flCommitRange)
		if err != nil {
			log.Fatal(err)
		}

		for _, commit := range c {
			fmt.Printf(" * %s %s ... ", commit.AbbreviatedCommit, commit.Subject)
			if err := ValidateCommit(commit, DefaultRules); err != nil {
				invalids = append(invalids, InvalidCommit{commit, err})
				fmt.Printf("FAIL: %s\n", err)
			} else {
				fmt.Println("PASS")
			}
		}
	} else if os.Getenv("TRAVIS_COMMIT_RANGE") != "" {
		c, err := GitCommits(os.Getenv("TRAVIS_COMMIT_RANGE"))
		if err != nil {
			log.Fatal(err)
		}

		for _, commit := range c {
			fmt.Printf(" * %s %s ... ", commit.AbbreviatedCommit, commit.Subject)
			if err := ValidateCommit(commit, DefaultRules); err != nil {
				invalids = append(invalids, InvalidCommit{commit, err})
				fmt.Printf("FAIL: %s\n", err)
			} else {
				fmt.Println("PASS")
			}
		}
	}
	if len(invalids) > 0 {
		fmt.Printf("%d issues to fix\n", len(invalids))
		os.Exit(1)
	}
}

type ValidateRule func(CommitEntry) error

func ValidateCommit(c CommitEntry, rules []ValidateRule) error {
	for _, r := range rules {
		if err := r(c); err != nil {
			return err
		}
	}
	return nil
}

type InvalidCommit struct {
	CommitEntry CommitEntry
	Error       error
}

type CommitEntry struct {
	Commit               string
	AbbreviatedCommit    string `json:"abbreviated_commit"`
	Tree                 string
	AbbreviatedTree      string `json:"abbreviated_tree"`
	Parent               string
	AbbreviatedParent    string `json:"abbreviated_parent"`
	Refs                 string
	Encoding             string
	Subject              string
	SanitizedSubjectLine string `json:"sanitized_subject_line"`
	Body                 string
	CommitNotes          string `json:"commit_notes"`
	VerificationFlag     string `json:"verification_flag"`
	ShortMsg             string
	Signer               string
	SignerKey            string `json:"signer_key"`
	Author               Person `json:"author,omitempty"`
	Commiter             Person `json:"commiter,omitempty"`
}

type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"` // this could maybe be an actual time.Time
}

var (
	prettyLogSubject = `--pretty=format:%s`
	prettyLogBody    = `--pretty=format:%b`
	prettyLogCommit  = `--pretty=format:%H`
	prettyLogFormat  = `--pretty=format:{"commit": "%H", "abbreviated_commit": "%h", "tree": "%T", "abbreviated_tree": "%t", "parent": "%P", "abbreviated_parent": "%p", "refs": "%D", "encoding": "%e", "sanitized_subject_line": "%f", "commit_notes": "%N", "verification_flag": "%G?", "signer": "%GS", "signer_key": "%GK", "author": { "name": "%aN", "email": "%aE", "date": "%aD" }, "commiter": { "name": "%cN", "email": "%cE", "date": "%cD" }}`
)

func GitLogCommit(commit string) (*CommitEntry, error) {
	buf := bytes.NewBuffer([]byte{})
	cmd := exec.Command("git", "log", "-1", prettyLogFormat, commit)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Println(strings.Join(cmd.Args, " "))
		return nil, err
	}
	c := CommitEntry{}
	output := buf.Bytes()
	if err := json.Unmarshal(output, &c); err != nil {
		fmt.Println(string(output))
		return nil, err
	}
	output, err := exec.Command("git", "log", "-1", prettyLogSubject, commit).Output()
	if err != nil {
		return nil, err
	}
	c.Subject = strings.TrimSpace(string(output))
	output, err = exec.Command("git", "log", "-1", prettyLogBody, commit).Output()
	if err != nil {
		return nil, err
	}
	c.Body = strings.TrimSpace(string(output))
	return &c, nil
}

func GitCommits(commitrange string) ([]CommitEntry, error) {
	output, err := exec.Command("git", "log", prettyLogCommit, commitrange).Output()
	if err != nil {
		return nil, err
	}
	commitHashes := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]CommitEntry, len(commitHashes))
	for i, commitHash := range commitHashes {
		c, err := GitLogCommit(commitHash)
		if err != nil {
			return commits, err
		}
		commits[i] = *c
	}
	return commits, nil
}

func GitFetchHeadCommit() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--verify", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
func GitHeadCommit() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--verify", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
