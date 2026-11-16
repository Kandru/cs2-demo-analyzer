package stats

// run.go orchestrates the scoring pipeline across every player:
//
//  1. Evaluate the lobby-independent detections for every player.
//  2. Append lobby-dependent detections (pre_fov_presence).
//  3. Lobby-relative normalize each detection.
//  4. Per player: combine via Bayesian log-odds, then apply modifiers in
//     registry order, clamp, and publish.
func evaluateScores(demoStats *DemoStats) {
	if demoStats == nil || len(demoStats.Players) == 0 {
		return
	}

	perPlayer := make(map[uint64][]Channel, len(demoStats.Players))
	for sid, ps := range demoStats.Players {
		if ps == nil {
			continue
		}
		perPlayer[sid] = evaluateChannelsForPlayer(ps)
	}

	ctx := buildLobbyContext(demoStats)
	for _, det := range detections {
		if det.LobbyEvaluate == nil {
			continue
		}
		for sid, ps := range demoStats.Players {
			if ps == nil {
				continue
			}
			perPlayer[sid] = append(perPlayer[sid], det.LobbyEvaluate(ctx, ps, sid))
		}
	}

	normalizeLobby(perPlayer)

	for sid, ps := range demoStats.Players {
		if ps == nil {
			continue
		}
		channels := perPlayer[sid]
		if channels == nil {
			channels = []Channel{}
		}

		st := &scoreState{
			combined:        bayesianCombine(channels),
			ps:              ps,
			channels:        channels,
			lobbyAsymmetric: ctx.IsAsymmetric(sid),
		}
		st.score = st.combined

		applied := make(map[string]Applied, len(modifiers))
		for _, m := range modifiers {
			applied[m.ID] = m.Apply(st)
		}
		if st.score > 100.0 {
			st.score = 100.0
		}

		publishScores(st, applied)
	}
}
