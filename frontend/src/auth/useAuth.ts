import { useState, useEffect, useCallback } from 'react';
import { CognitoUserSession } from 'amazon-cognito-identity-js';
import * as cognito from './cognitoClient';

interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  userId: string | null;
  email: string | null;
  error: string | null;
}

export function useAuth() {
  const [state, setState] = useState<AuthState>({
    isAuthenticated: false,
    isLoading: true,
    userId: null,
    email: null,
    error: null,
  });

  // Check existing session on mount
  useEffect(() => {
    checkSession();
  }, []);

  const checkSession = async () => {
    try {
      const session = await cognito.getCurrentSession();
      if (session && session.isValid()) {
        const payload = session.getIdToken().decodePayload();
        setState({
          isAuthenticated: true,
          isLoading: false,
          userId: payload.sub,
          email: payload.email,
          error: null,
        });
      } else {
        setState(prev => ({ ...prev, isLoading: false }));
      }
    } catch {
      setState(prev => ({ ...prev, isLoading: false }));
    }
  };

  const login = useCallback(async (email: string, password: string) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }));
    try {
      const session: CognitoUserSession = await cognito.signIn(email, password);
      const payload = session.getIdToken().decodePayload();
      setState({
        isAuthenticated: true,
        isLoading: false,
        userId: payload.sub,
        email: payload.email,
        error: null,
      });
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Login failed';
      setState(prev => ({ ...prev, isLoading: false, error: message }));
      throw err;
    }
  }, []);

  const register = useCallback(async (email: string, password: string, nickname: string) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }));
    try {
      await cognito.signUp(email, password, nickname);
      setState(prev => ({ ...prev, isLoading: false }));
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Registration failed';
      setState(prev => ({ ...prev, isLoading: false, error: message }));
      throw err;
    }
  }, []);

  const confirmRegistration = useCallback(async (email: string, code: string) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }));
    try {
      await cognito.confirmSignUp(email, code);
      setState(prev => ({ ...prev, isLoading: false }));
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Confirmation failed';
      setState(prev => ({ ...prev, isLoading: false, error: message }));
      throw err;
    }
  }, []);

  const logout = useCallback(async () => {
    await cognito.signOut();
    setState({
      isAuthenticated: false,
      isLoading: false,
      userId: null,
      email: null,
      error: null,
    });
  }, []);

  const getToken = useCallback(async (): Promise<string | null> => {
    return cognito.getIdToken();
  }, []);

  return {
    ...state,
    login,
    register,
    confirmRegistration,
    logout,
    getToken,
  };
}
