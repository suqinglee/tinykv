package raft

import (
	pb "github.com/pingcap-incubator/tinykv/proto/pkg/eraftpb"
	"sort"
)

func (r *Raft) handleHup(m pb.Message) {
	if r.State != StateFollower && r.State != StateCandidate {return}
	r.becomeCandidate()
	if len(r.Prs) == 1 {r.becomeLeader()}
	for id := range r.Prs {
		if id != r.id {r.sendRequestVote(id)}
	}
}

func (r *Raft) handleBeat(m pb.Message) {
	if r.State != StateLeader {return}
	for id := range r.Prs {
		if id != r.id {r.sendHeartbeat(id)}
	}
	r.heartbeatElapsed = 0
}

func (r *Raft) handlePropose(m pb.Message) {
	if r.State != StateLeader {return}
	for i, e := range m.Entries {
		e.Index = uint64(i) + r.RaftLog.LastIndex() + 1
		e.Term = r.Term
		r.RaftLog.entries = append(r.RaftLog.entries, *e)
	}
	r.Prs[r.id] = &Progress{
		Next: r.RaftLog.LastIndex() + 1,
		Match: r.RaftLog.LastIndex(),
	}
	r.bcastAppend()
}

func (r *Raft) handleAppendEntries(m pb.Message) {
	for {
		if m.Term < r.Term {
			break
		}
		if m.Term >= r.Term {
			r.becomeFollower(m.Term, m.From)
		}
		logTerm, err := r.RaftLog.Term(m.Index)
		if err != nil || logTerm != m.LogTerm {
			break
		}
		for i, e := range m.Entries {
			j := m.Index + uint64(i) + 1 - r.RaftLog.FirstIndex()
			if j >= uint64(len(r.RaftLog.entries)) {
				r.RaftLog.entries = append(r.RaftLog.entries, *e)
			} else {
				if r.RaftLog.entries[j].Term != e.Term {
					r.RaftLog.stabled = min(j+r.RaftLog.FirstIndex()-1, r.RaftLog.stabled)
					r.RaftLog.entries = append(r.RaftLog.entries[:j], *e)
				} else {
					r.RaftLog.entries[j] = *e
				}
			}
		}
		if m.Commit > r.RaftLog.committed {
			r.RaftLog.committed = min(m.Commit, r.RaftLog.LastIndex())
		}
		r.sendAppendResponse(m, false)
		return
	}
	r.sendAppendResponse(m, true)
}

func (r *Raft) handleAppendEntriesResponse(m pb.Message) {
	if r.State != StateLeader {return}
	if m.Reject {
		r.Prs[m.From].Next = max(1, r.Prs[m.From].Next - 1)
		r.sendAppend(m.From)
	} else {
		r.Prs[m.From] = &Progress{
			Match: m.Index,
			Next: m.Index + 1,
		}
		var match []uint64
		for _, prs := range r.Prs {
			match = append(match, prs.Match)
		}
		sort.Slice(match, func(i, j int) bool { return match[i] < match[j] })
		majority := match[(len(r.Prs)-1)/2]
		logTerm, _ := r.RaftLog.Term(majority)
		if majority > r.RaftLog.committed && logTerm == r.Term {
			r.RaftLog.committed = majority
			r.bcastAppend()
		}
	}
}

func (r *Raft) handleRequestVote(m pb.Message) {
	for {
		if m.Term < r.Term {
			break
		}
		if m.Term > r.Term {
			r.becomeFollower(m.Term, None)
		}
		if r.Vote != 0 && r.Vote != m.From {
			break
		}
		logTerm, index := r.RaftLog.LastTerm(), r.RaftLog.LastIndex()
		if m.LogTerm < logTerm || m.LogTerm == logTerm && m.Index < index {
			break
		}
		r.Vote = m.From
		r.sendRequestVoteResponse(m, false)
		return
	}
	r.sendRequestVoteResponse(m, true)
}

func (r *Raft) handleRequestVoteResponse(m pb.Message) {
	if r.State != StateCandidate {return}
	r.votes[m.From] = !m.Reject
	good, bad := 0, 0
	for id := range r.votes {
		if r.votes[id] {
			good++
		} else {
			bad++
		}
	}
	if good > len(r.Prs) / 2 {
		r.becomeLeader()
	} else if bad > len(r.Prs) / 2 {
		r.becomeFollower(r.Term, None)
	}
}

func (r *Raft) handleSnapshot(m pb.Message) {

}

func (r *Raft) handleHeartbeat(m pb.Message) {
	for {
		if m.Term < r.Term {
			break
		}
		if m.Term >= r.Term {
			r.becomeFollower(m.Term, m.From)
		}
		r.sendHeartbeatResponse(m, false)
		return
	}
	r.sendHeartbeatResponse(m, true)
}

func (r *Raft) handleHeartbeatResponse(m pb.Message) {
	if r.State != StateLeader {return}
	if m.Reject {
		r.becomeFollower(m.Term, m.From)
	}
}

func (r *Raft) handleTransferLeader(m pb.Message) {

}

func (r *Raft) handleTimeoutNow(m pb.Message) {

}