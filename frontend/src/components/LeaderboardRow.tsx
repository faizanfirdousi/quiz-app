interface LeaderboardRowProps {
  rank: number;
  nickname: string;
  score: number;
}

const RANK_EMOJIS: Record<number, string> = {
  1: '🥇',
  2: '🥈',
  3: '🥉',
};

export default function LeaderboardRow({ rank, nickname, score }: LeaderboardRowProps) {
  return (
    <div className="leaderboard-row">
      <div className="flex items-center gap-3">
        <span className="w-8 text-center font-display font-bold text-lg">
          {RANK_EMOJIS[rank] || rank}
        </span>
        <span className="font-medium truncate max-w-[200px]">{nickname}</span>
      </div>
      <span className="font-display font-bold text-purple-400">
        {score.toLocaleString()}
      </span>
    </div>
  );
}
