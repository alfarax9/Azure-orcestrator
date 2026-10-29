import { Eye, EyeOff } from "lucide-react";
import { useState, type FormEvent } from "react";
import { Button, Form, Input, Label, TextField } from "react-aria-components";
import { ApiError, type Api } from "../daemon/api";
import "./PairScreen.css";

function defaultDeviceName(): string {
  const platform = navigator.platform || "Browser";
  return `Web on ${platform}`;
}

/** What a failed pairing says, in words that tell the person what to do next. */
function failure(e: unknown): string {
  if (e instanceof ApiError) {
    if (e.status === 401) return "That token didn't work. It may have expired or been used already: run astrad pair on the server for a new one.";
    if (e.status === 0) return "Couldn't reach the daemon. Check that astrad serve is running and that this page is open through its address or tunnel.";
    return e.message;
  }
  return "Pairing failed. Try again.";
}

/**
 * Pairing (F1-05): trade the one-time token that `astrad pair` printed for this browser's own
 * token. The token field hides what is typed or pasted, and nothing is kept after it is used.
 */
export function PairScreen({ api, notice, onPaired }: { api: Pick<Api, "pair">; notice?: string; onPaired: (deviceToken: string) => void }) {
  const [token, setToken] = useState("");
  const [name, setName] = useState(defaultDeviceName);
  const [shown, setShown] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const deviceToken = await api.pair(token.trim(), name.trim());
      setToken("");
      onPaired(deviceToken);
    } catch (err) {
      setError(failure(err));
      setBusy(false);
    }
  }

  return (
    <main className="pair" id="main" tabIndex={-1}>
      <div className="pair-panel">
        <h1>Pair this device</h1>
        {notice && <p className="pair-notice">{notice}</p>}
        <p className="pair-help">
          Run <code>astrad pair</code> on the server and paste the token it prints. A token works once and expires after 10 minutes.
        </p>
        <Form className="pair-form" onSubmit={submit} validationBehavior="native">
          <TextField className="field" isRequired value={token} onChange={setToken} isInvalid={error !== null} autoComplete="off">
            <Label>Pairing token</Label>
            <div className="field-with-action">
              <Input type={shown ? "text" : "password"} spellCheck={false} autoCapitalize="off" autoCorrect="off" />
              <Button className="field-action" aria-label={shown ? "Hide token" : "Show token"} onPress={() => setShown(!shown)}>
                {shown ? <EyeOff size={16} strokeWidth={1.5} aria-hidden="true" /> : <Eye size={16} strokeWidth={1.5} aria-hidden="true" />}
              </Button>
            </div>
          </TextField>
          <TextField className="field" isRequired value={name} onChange={setName}>
            <Label>Device name</Label>
            <Input />
            <p className="field-help">Shown in the list of paired devices, so you can cut this one off later.</p>
          </TextField>
          <Button type="submit" className="button-primary" isDisabled={busy || token.trim() === ""}>
            {busy ? "Pairing…" : "Pair device"}
          </Button>
        </Form>
        {error && (
          <p className="field-error" role="alert">
            {error}
          </p>
        )}
      </div>
    </main>
  );
}
