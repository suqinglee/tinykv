package raft

import (
	pb "github.com/pingcap-incubator/tinykv/proto/pkg/eraftpb"
)

// sendAppend sends an append RPC with new entries (if any) and the
// current commit index to the given peer. Returns true if a message was sent.
func (r *Raft) sendAppend(to uint64) {
	msg := pb.Message{
		MsgType: pb.MessageType_MsgAppend,
		To: to,
		From: r.id,
		Term: r.Term,
		Index: r.Prs[to].Next - 1,
		Commit: r.RaftLog.committed,
	}
	msg.LogTerm, _ = r.RaftLog.Term(msg.Index)
	for i := r.Prs[to].Next; i <= r.RaftLog.LastIndex(); i++ {
		j := i - r.RaftLog.FirstIndex()
		msg.Entries = append(msg.Entries, &r.RaftLog.entries[j])
	}
	r.msgs = append(r.msgs, msg)
}

func (r *Raft) sendAppendResponse(m pb.Message, reject bool) {
	r.msgs = append(r.msgs, pb.Message{
		MsgType: pb.MessageType_MsgAppendResponse,
		To: m.From,
		From: r.id,
		Term: r.Term,
		Index: m.Index + uint64(len(m.Entries)),
		Reject: reject,
	})
}

func (r *Raft) sendRequestVote(to uint64) {
	r.msgs = append(r.msgs, pb.Message{
		MsgType: pb.MessageType_MsgRequestVote,
		To: to,
		From: r.id,
		Term: r.Term,
		Index: r.RaftLog.LastIndex(),
		LogTerm: r.RaftLog.LastTerm(),
	})
}

func (r *Raft) sendRequestVoteResponse(m pb.Message, reject bool) {
	r.msgs = append(r.msgs, pb.Message{
		MsgType: pb.MessageType_MsgRequestVoteResponse,
		To: m.From,
		From: r.id,
		Term: r.Term,
		Reject: reject,
	})
}

// sendHeartbeat sends a heartbeat RPC to the given peer.
func (r *Raft) sendHeartbeat(to uint64) {
	r.msgs = append(r.msgs, pb.Message{
		MsgType: pb.MessageType_MsgHeartbeat,
		To: to,
		From: r.id,
		Term: r.Term,
		Commit: r.RaftLog.committed,
	})
}

func (r *Raft) sendHeartbeatResponse(m pb.Message, reject bool) {
	r.msgs = append(r.msgs, pb.Message{
		MsgType: pb.MessageType_MsgHeartbeatResponse,
		To: m.From,
		From: r.id,
		Term: r.Term,
		Reject: reject,
		Commit: r.RaftLog.committed,
	})
}
