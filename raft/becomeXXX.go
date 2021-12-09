package raft

import (
	pb "github.com/pingcap-incubator/tinykv/proto/pkg/eraftpb"
	"math/rand"
)

func (r *Raft) becomeFollower(term uint64, lead uint64) {
	r.State = StateFollower
	r.Lead = lead
	r.Term = term
	r.Vote = None
	r.votes = make(map[uint64]bool)
	r.electionElapsed = -rand.Intn(r.electionTimeout)
}

func (r *Raft) becomeCandidate() {
	r.State = StateCandidate
	r.Lead = None
	r.Term++
	r.Vote = r.id
	r.votes = map[uint64]bool{r.id: true}
	r.electionElapsed = -rand.Intn(r.electionTimeout)
}

func (r *Raft) becomeLeader() {
	// NOTE: Leader should propose a noop entry on its term
	r.State = StateLeader
	r.Lead = r.id
	r.heartbeatElapsed = 0

	for id := range r.Prs {
		r.Prs[id] = &Progress{
			Next: r.RaftLog.LastIndex() + 1,
			Match: 0,
		}
	}
	r.RaftLog.entries = append(r.RaftLog.entries, pb.Entry{
		Term: r.Term,
		Index: r.RaftLog.LastIndex() + 1,
	})
	r.Prs[r.id] = &Progress{
		Next: r.RaftLog.LastIndex() + 1,
		Match: r.RaftLog.LastIndex(),
	}
	r.bcastAppend()
}