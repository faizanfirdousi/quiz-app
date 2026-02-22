import { motion } from 'framer-motion';

interface CountdownTimerProps {
  timeRemaining: number; // ms
  totalTime: number;     // ms
}

export default function CountdownTimer({ timeRemaining, totalTime }: CountdownTimerProps) {
  const seconds = Math.ceil(timeRemaining / 1000);
  const progress = totalTime > 0 ? timeRemaining / totalTime : 0;
  const radius = 24;
  const circumference = 2 * Math.PI * radius;
  const dashOffset = circumference * (1 - progress);

  const color = progress > 0.5 ? '#26890C' : progress > 0.25 ? '#D89E00' : '#E21B3C';

  return (
    <div className="relative inline-flex items-center justify-center">
      <svg width="64" height="64" className="-rotate-90">
        {/* Background circle */}
        <circle
          cx="32"
          cy="32"
          r={radius}
          fill="none"
          stroke="rgba(255,255,255,0.1)"
          strokeWidth="4"
        />
        {/* Progress circle */}
        <motion.circle
          cx="32"
          cy="32"
          r={radius}
          fill="none"
          stroke={color}
          strokeWidth="4"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={dashOffset}
        />
      </svg>
      <span className="absolute text-xl font-display font-bold" style={{ color }}>
        {seconds}
      </span>
    </div>
  );
}
