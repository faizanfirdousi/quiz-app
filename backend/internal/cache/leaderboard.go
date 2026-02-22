package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"kahootclone/internal/models"
	"kahootclone/internal/observability"
)

const leaderboardKeyPrefix = "leaderboard:"

func leaderboardKey(sessionID string) string {
	return leaderboardKeyPrefix + sessionID
}

// nicknameKey stores the mapping from userId to nickname for leaderboard display.
const nicknameKeyPrefix = "nickname:"

func nicknameKey(sessionID string) string {
	return nicknameKeyPrefix + sessionID
}

// UpsertScore adds or updates a player's score in the leaderboard.
func (r *RedisClient) UpsertScore(ctx context.Context, sessionID, userID string, score float64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "upserting score", "sessionId", sessionID, "userId", userID, "score", score)

	return r.Client.ZAdd(ctx, leaderboardKey(sessionID), redis.Z{
		Score:  score,
		Member: userID,
	}).Err()
}

// IncrementScore atomically increments a player's score.
func (r *RedisClient) IncrementScore(ctx context.Context, sessionID, userID string, delta float64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "incrementing score", "sessionId", sessionID, "userId", userID, "delta", delta)

	return r.Client.ZIncrBy(ctx, leaderboardKey(sessionID), delta, userID).Err()
}

// SetNickname stores a user's nickname for leaderboard display.
func (r *RedisClient) SetNickname(ctx context.Context, sessionID, userID, nickname string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.Client.HSet(ctx, nicknameKey(sessionID), userID, nickname).Err()
}

// GetTopN returns the top N players with scores, sorted descending.
func (r *RedisClient) GetTopN(ctx context.Context, sessionID string, n int) ([]models.PlayerScore, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "getting top N", "sessionId", sessionID, "n", n)

	results, err := r.Client.ZRevRangeWithScores(ctx, leaderboardKey(sessionID), 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}

	// Fetch all nicknames for this session
	nicknames, _ := r.Client.HGetAll(ctx, nicknameKey(sessionID)).Result()

	scores := make([]models.PlayerScore, len(results))
	for i, z := range results {
		userID := z.Member.(string)
		nickname := nicknames[userID]
		if nickname == "" {
			nickname = userID[:8] // fallback
		}
		scores[i] = models.PlayerScore{
			UserID:   userID,
			Nickname: nickname,
			Score:    z.Score,
			Rank:     int64(i + 1),
		}
	}
	return scores, nil
}

// GetPlayerRank returns a player's rank (1-indexed from top).
func (r *RedisClient) GetPlayerRank(ctx context.Context, sessionID, userID string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rank, err := r.Client.ZRevRank(ctx, leaderboardKey(sessionID), userID).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, fmt.Errorf("player %s not found in leaderboard", userID)
		}
		return -1, err
	}
	return rank + 1, nil // Convert 0-indexed to 1-indexed
}

// GetPlayerScore returns a player's current score.
func (r *RedisClient) GetPlayerScore(ctx context.Context, sessionID, userID string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	score, err := r.Client.ZScore(ctx, leaderboardKey(sessionID), userID).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return score, nil
}

// GetPlayerCount returns the total number of players in the leaderboard.
func (r *RedisClient) GetPlayerCount(ctx context.Context, sessionID string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.Client.ZCard(ctx, leaderboardKey(sessionID)).Result()
}

// DeleteSession removes all leaderboard data for a session.
func (r *RedisClient) DeleteSession(ctx context.Context, sessionID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "deleting session leaderboard", "sessionId", sessionID)

	pipe := r.Client.Pipeline()
	pipe.Del(ctx, leaderboardKey(sessionID))
	pipe.Del(ctx, nicknameKey(sessionID))
	_, err := pipe.Exec(ctx)
	return err
}
