export interface VpnProfile {
  id: string;
  filename: string;
  country_code: string;
  country: string;
  city: string;
  is_virtual: boolean;
  endpoints: string[];
  full_path: string;
}

export interface AppConfig {
  username?: string;
  password?: string;
  favorites: string[];
  last_selected_profile?: string;
  last_selected_endpoint?: string;
  protocol?: "udp" | "tcp";
  server_list_source?: "downloaded" | "custom";
  last_server_list_update?: number;
  custom_udp_path?: string;
  custom_tcp_path?: string;
}

export interface ServerListInfo {
  protocol: "udp" | "tcp";
  server_list_source: "downloaded" | "custom";
  last_server_list_update: number;
  udp_count: number;
  tcp_count: number;
  has_profiles: boolean;
  custom_udp_path?: string;
  custom_tcp_path?: string;
}

export type ConnectionState =
  | "Disconnected"
  | "Connecting"
  | "Connected"
  | "Disconnecting"
  | "Error";

export interface ConnectionStatusData {
  profile_id?: string;
  endpoint?: string;
  resolved_endpoint?: string;
  assigned_ip?: string;
  public_ip?: string;
  connected_at?: number;
  error?: string;
}

export interface ConnectionStatus {
  status: ConnectionState;
  data?: ConnectionStatusData;
}

