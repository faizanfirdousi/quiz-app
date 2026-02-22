import { useEffect, useRef, useCallback, useState } from 'react';
import { getIdToken } from '../auth/cognitoClient';
import { ConnectionStatus } from '../types';
import {
  WS_TYPES,
  type PlayerJoinedPayload,
  type GameStartedPayload,
  type QuestionPayload,
  type AnswerResultPayload,
  type QuestionEndedPayload,
  type LeaderboardUpdatePayload,
  type GameOverPayload,
  type ErrorPayload,
} from './eventTypes';
import { useGameStore } from '../store/gameStore';
import { useLeaderboardStore } from '../store/leaderboardStore';

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';
const MAX_RETRIES = 5;
const HEARTBEAT_INTERVAL = 30000;

export function useGameSocket(sessionId: string | null, role: 'HOST' | 'PLAYER' = 'PLAYER') {
  const wsRef = useRef<WebSocket | null>(null);
  const retriesRef = useRef(0);
  const heartbeatRef = useRef<ReturnType<typeof setInterval>>();
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('disconnected');

  const gameStore = useGameStore();
  const leaderboardStore = useLeaderboardStore();

  const connect = useCallback(async () => {
    if (!sessionId) return;

    setConnectionStatus('connecting');

    let token: string | null = null;
    try {
      token = await getIdToken();
    } catch {
      console.warn('No auth token available for WS');
    }

    const params = new URLSearchParams({ sessionId, role });
    if (token) params.set('token', token);

    const ws = new WebSocket(`${WS_URL}?${params.toString()}`);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnectionStatus('connected');
      retriesRef.current = 0;

      // Start heartbeat
      heartbeatRef.current = setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ action: 'ping', data: {} }));
        }
      }, HEARTBEAT_INTERVAL);
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        handleMessage(message);
      } catch (err) {
        console.error('Failed to parse WS message:', err);
      }
    };

    ws.onclose = () => {
      setConnectionStatus('disconnected');
      clearInterval(heartbeatRef.current);

      // Auto-reconnect with exponential backoff
      if (retriesRef.current < MAX_RETRIES) {
        const delay = Math.pow(2, retriesRef.current) * 1000; // 1s, 2s, 4s, 8s, 16s
        retriesRef.current++;
        setTimeout(() => connect(), delay);
      }
    };

    ws.onerror = () => {
      setConnectionStatus('error');
    };
  }, [sessionId, role]);

  const handleMessage = useCallback((message: { type: string; payload: unknown }) => {
    switch (message.type) {
      case WS_TYPES.PLAYER_JOINED: {
        const payload = message.payload as PlayerJoinedPayload;
        gameStore.setPlayerCount(payload.playerCount);
        break;
      }
      case WS_TYPES.GAME_STARTED: {
        const payload = message.payload as GameStartedPayload;
        gameStore.setGameStatus('ACTIVE');
        gameStore.setTotalQuestions(payload.totalQuestions);
        break;
      }
      case WS_TYPES.QUESTION: {
        const payload = message.payload as QuestionPayload;
        gameStore.setQuestion(payload);
        gameStore.startCountdown(payload.timeLimitMs);
        break;
      }
      case WS_TYPES.ANSWER_RESULT: {
        const payload = message.payload as AnswerResultPayload;
        gameStore.setAnswerResult(payload);
        leaderboardStore.updateMyScore(payload.totalScore, payload.rank);
        break;
      }
      case WS_TYPES.QUESTION_ENDED: {
        const payload = message.payload as QuestionEndedPayload;
        leaderboardStore.setLeaderboard(payload.leaderboard);
        break;
      }
      case WS_TYPES.LEADERBOARD_UPDATE: {
        const payload = message.payload as LeaderboardUpdatePayload;
        leaderboardStore.setLeaderboard(payload.leaderboard);
        break;
      }
      case WS_TYPES.GAME_OVER: {
        const payload = message.payload as GameOverPayload;
        gameStore.setGameStatus('FINISHED');
        leaderboardStore.setLeaderboard(payload.finalLeaderboard);
        break;
      }
      case WS_TYPES.ERROR: {
        const payload = message.payload as ErrorPayload;
        console.error(`WS Error [${payload.code}]: ${payload.message}`);
        break;
      }
    }
  }, [gameStore, leaderboardStore]);

  const sendMessage = useCallback((action: string, data: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ action, data }));
    } else {
      console.error('WebSocket is not connected');
    }
  }, []);

  const disconnect = useCallback(() => {
    retriesRef.current = MAX_RETRIES; // Prevent reconnect
    clearInterval(heartbeatRef.current);
    wsRef.current?.close();
    setConnectionStatus('disconnected');
  }, []);

  // Connect when sessionId changes
  useEffect(() => {
    if (sessionId) {
      connect();
    }
    return () => {
      disconnect();
    };
  }, [sessionId]);

  return {
    connectionStatus,
    sendMessage,
    disconnect,
  };
}
