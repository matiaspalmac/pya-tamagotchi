export type WSMessage = {
  type: string;
  data: Record<string, unknown>;
};

export function connectWS(token: string, onMessage: (msg: WSMessage) => void): WebSocket {
  const apiURL = (import.meta.env.VITE_API_URL as string | undefined) ?? "";
  let wsBase: string;
  if (apiURL) {
    wsBase = apiURL.replace(/^http/, "ws");
  } else {
    const proto = location.protocol === "https:" ? "wss" : "ws";
    wsBase = `${proto}://${location.host}`;
  }
  const ws = new WebSocket(`${wsBase}/ws?token=${encodeURIComponent(token)}`);
  ws.onmessage = (e) => {
    try {
      onMessage(JSON.parse(e.data));
    } catch {}
  };
  return ws;
}
