package wrf

// React Query hook templates

const queryHookTemplate = `export function {{.HookName}}({{.Params}}) {
  return useQuery({
    queryKey: {{.QueryKey}},
    queryFn: async () => {
      const result = await window.go.{{.ServiceName}}.{{.MethodName}}({{.ParamNames}});
      return result;
    },
    {{- if .Annotation.Cache}}
    cacheTime: {{.Annotation.Cache | formatDuration}},
    {{- end}}
    {{- if .Annotation.StaleTime}}
    staleTime: {{.Annotation.StaleTime | formatDuration}},
    {{- end}}
    {{- if .Annotation.Events}}
    // Auto-invalidation on events
    onMount: (queryClient) => {
      const unsubscribes = [
        {{- range .Annotation.Events}}
        EventsOn('{{.}}', () => {
          queryClient.invalidateQueries({ queryKey: {{$.QueryKey}} });
        }),
        {{- end}}
      ];
      
      return () => {
        unsubscribes.forEach(fn => fn());
      };
    },
    {{- end}}
    {{- if not .Annotation.Enabled}}
    enabled: false,
    {{- end}}
  });
}`

const mutationHookTemplate = `export function {{.HookName}}() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({{.Params}}) => {
      return window.go.{{.ServiceName}}.{{.MethodName}}({{.ParamNames}});
    },
    {{- if .Annotation.Optimistic}}
    onMutate: async (variables) => {
      // Optimistic update
      {{- range .InvalidateQueries}}
      await queryClient.cancelQueries({ queryKey: {{.}} });
      {{- end}}
      
      const previousData = queryClient.getQueryData({{.OptimisticKey}});
      
      queryClient.setQueryData({{.OptimisticKey}}, (old) => {
        // Apply optimistic update
        return { ...old, ...variables };
      });
      
      return { previousData };
    },
    onError: (err, variables, context) => {
      // Rollback on error
      if (context?.previousData) {
        queryClient.setQueryData({{.OptimisticKey}}, context.previousData);
      }
    },
    {{- end}}
    onSuccess: (data, variables) => {
      {{- range .InvalidateQueries}}
      queryClient.invalidateQueries({ queryKey: {{.}} });
      {{- end}}
      {{- range .Annotation.Events}}
      // Emit events for real-time sync
      EventsEmit('{{.}}', data);
      {{- end}}
      {{- if .Annotation.OnSuccess}}
      // Custom success handler
      if (typeof {{.Annotation.OnSuccess}} === 'function') {
        {{.Annotation.OnSuccess}}(data, variables);
      }
      {{- end}}
    },
    {{- if .Annotation.RetryCount}}
    retry: {{.Annotation.RetryCount}},
    {{- end}}
  });
}`

const subscriptionHookTemplate = `export function {{.HookName}}() {
  const [data, setData] = useState<{{.ReturnType}} | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  
  useEffect(() => {
    {{- if .Annotation.AutoStart}}
    let unsubscribe: (() => void) | null = null;
    
    const connect = () => {
      setIsConnected(true);
      setError(null);
      
      unsubscribe = EventsOn('{{.Annotation.EventName}}', (eventData: {{.ReturnType}}) => {
        setData(eventData);
      });
    };
    
    connect();
    
    return () => {
      if (unsubscribe) {
        unsubscribe();
      }
      setIsConnected(false);
    };
    {{- else}}
    // Manual subscription - call connect() to start
    {{- end}}
  }, []);
  
  const connect = useCallback(() => {
    if (isConnected) return;
    
    setIsConnected(true);
    setError(null);
    
    return EventsOn('{{.Annotation.EventName}}', (eventData: {{.ReturnType}}) => {
      setData(eventData);
    });
  }, [isConnected]);
  
  const disconnect = useCallback(() => {
    setIsConnected(false);
    setData(null);
  }, []);
  
  return { 
    data, 
    isConnected, 
    error, 
    connect, 
    disconnect 
  };
}`

const basicHookTemplate = `export function {{.HookName}}() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  
  const execute = useCallback(async ({{.Params}}) => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await window.go.{{.ServiceName}}.{{.MethodName}}({{.ParamNames}});
      return result;
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);
  
  return { execute, loading, error };
}`

// SWR hook templates (alternative to React Query)

const swrQueryHookTemplate = `export function {{.HookName}}({{.Params}}) {
  const { data, error, mutate } = useSWR(
    {{.QueryKey}},
    async () => {
      const result = await window.go.{{.ServiceName}}.{{.MethodName}}({{.ParamNames}});
      return result;
    },
    {
      {{- if .Annotation.Cache}}
      dedupingInterval: {{.Annotation.Cache | formatDuration}},
      {{- end}}
      {{- if .Annotation.StaleTime}}
      focusThrottleInterval: {{.Annotation.StaleTime | formatDuration}},
      {{- end}}
      {{- if .Annotation.Events}}
      // Auto-revalidation on events
      onMount: () => {
        const unsubscribes = [
          {{- range .Annotation.Events}}
          EventsOn('{{.}}', () => {
            mutate();
          }),
          {{- end}}
        ];
        
        return () => {
          unsubscribes.forEach(fn => fn());
        };
      },
      {{- end}}
    }
  );
  
  return {
    data,
    error,
    isLoading: !error && !data,
    mutate
  };
}`

const swrMutationHookTemplate = `export function {{.HookName}}() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  
  const trigger = useCallback(async ({{.Params}}) => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await window.go.{{.ServiceName}}.{{.MethodName}}({{.ParamNames}});
      
      // Revalidate related queries
      {{- range .InvalidateQueries}}
      mutate({{.}});
      {{- end}}
      
      // Emit events
      {{- range .Annotation.Events}}
      EventsEmit('{{.}}', result);
      {{- end}}
      
      return result;
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);
  
  return { trigger, loading, error };
}`

// Configuration examples and documentation templates

const configExampleTemplate = `{
  "plugins": {
    "wrf": {
      "enabled": true,
      "config": {
        "go": {
          "sourcePaths": ["./app/**/*.go"],
          "watchMode": true,
          "buildTags": ["wrf", "desktop"]
        },
        "generation": {
          "hooks": {
            "output": "frontend/src/hooks/generated",
            "framework": "react-query",
            "includeComments": true,
            "cacheStrategies": true
          },
          "types": {
            "output": "frontend/src/types/generated.ts",
            "validation": "zod",
            "includeSchemas": true,
            "nullableFields": true
          },
          "events": {
            "output": "frontend/src/events/generated.ts",
            "typed": true
          }
        },
        "development": {
          "hmr": true,
          "devtools": true,
          "eventDebugging": true
        }
      }
    }
  }
}`

const usageExampleTemplate = `// Example Go service with WRF annotations

package app

// UserService manages user operations
// @wrf:service
type UserService struct {
	db *sql.DB
}

// GetUser retrieves a user by ID
// @wrf:query(cache: "5m", events: ["user:updated", "user:deleted"])
func (u *UserService) GetUser(ctx context.Context, id int) (*User, error) {
	// Implementation
}

// UpdateUser updates user information
// @wrf:mutation(optimistic: true, invalidate: ["UserService:GetUser"], events: ["user:updated"])
func (u *UserService) UpdateUser(ctx context.Context, user *User) (*User, error) {
	// Implementation
}

// WatchUserStatus subscribes to user status changes
// @wrf:subscription(eventName: "user:status", autoStart: true)
func (u *UserService) WatchUserStatus(ctx context.Context, userID int) error {
	// Implementation
}`

const reactUsageTemplate = `// Generated hooks usage in React components

import { useGetUser, useUpdateUser, useWatchUserStatus } from '@/hooks/generated';

function UserProfile({ userId }: { userId: number }) {
  // Query hook with caching and auto-invalidation
  const { data: user, isLoading, error } = useGetUser(userId);
  
  // Mutation hook with optimistic updates
  const { mutate: updateUser, isLoading: isUpdating } = useUpdateUser();
  
  // Subscription hook for real-time updates
  const { data: status, isConnected } = useWatchUserStatus();
  
  const handleUpdate = (userData: Partial<User>) => {
    updateUser({ ...user, ...userData });
  };
  
  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;
  
  return (
    <div>
      <h1>{user?.name}</h1>
      <p>Status: {status?.online ? 'Online' : 'Offline'}</p>
      <button onClick={() => handleUpdate({ name: 'New Name' })}>
        Update Name
      </button>
    </div>
  );
}`