import { useEffect, useMemo, useRef, useState } from "react";
import {
  useAddInterviewMessage,
  useCreateInterviewSession,
  useEndInterviewSession,
  useInterviewSession,
  useRunExtraction,
} from "../api/hooks";
import type { InterviewExtraction, InterviewMessage } from "../api/types";

// ponytail: Web Speech API types aren't in lib.dom defaults; declare the
// minimal surface we use. Ceiling: if TS ships these in a future lib, drop this.
interface SpeechRecognitionEvent extends Event {
  results: { length: number; [i: number]: { 0: { transcript: string }; isFinal: boolean } };
}
interface SpeechRecognitionLike {
  continuous: boolean;
  interimResults: boolean;
  lang: string;
  onresult: ((e: SpeechRecognitionEvent) => void) | null;
  onend: (() => void) | null;
  onerror: (() => void) | null;
  start: () => void;
  stop: () => void;
}
function getSpeechRecognitionCtor(): (new () => SpeechRecognitionLike) | null {
  const w = window as unknown as {
    SpeechRecognition?: new () => SpeechRecognitionLike;
    webkitSpeechRecognition?: new () => SpeechRecognitionLike;
  };
  return w.SpeechRecognition ?? w.webkitSpeechRecognition ?? null;
}

const AGENT_META: Record<string, { icon: string; label: string; color: string }> = {
  relationship_stage: { icon: "💍", label: "Relationship Stage", color: "#7c5cff" },
  wedding_timeline: { icon: "📅", label: "Wedding Timeline", color: "#4f8cff" },
  vendor_interest: { icon: "💐", label: "Vendor Interest", color: "#2dd4a8" },
  location: { icon: "📍", label: "Location", color: "#f3813f" },
  budget: { icon: "💰", label: "Budget", color: "#f87171" },
};

function speak(text: string) {
  if (!("speechSynthesis" in window)) return;
  const u = new SpeechSynthesisUtterance(text);
  u.rate = 1.0;
  u.pitch = 1.0;
  window.speechSynthesis.speak(u);
}

function relativeTime(iso: string): string {
  const s = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60);
  if (m < 60) return `${m}m`;
  const h = Math.floor(m / 60);
  return `${h}h`;
}

interface ChatColumnProps {
  coupleLabel: string;
  couple: "A" | "B";
  messages: InterviewMessage[];
  onSend: (speaker: string, text: string) => void;
  disabled: boolean;
  ttsEnabled: boolean;
}

function ChatColumn({ coupleLabel, couple, messages, onSend, disabled, ttsEnabled }: ChatColumnProps) {
  const [text, setText] = useState("");
  const [speaker, setSpeaker] = useState<"1" | "2">("1");
  const [listening, setListening] = useState(false);
  const [interim, setInterim] = useState("");
  const recRef = useRef<SpeechRecognitionLike | null>(null);
  const scrollRef = useRef<HTMLDivElement | null>(null);
  const seenIds = useRef<Set<string>>(new Set());

  const speakerCode = `${couple.toLowerCase()}${speaker}`; // a1/a2/b1/b2
  const personLabel = speaker === "1" ? "Person 1" : "Person 2";

  // Auto-scroll on new messages.
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: "smooth" });
  }, [messages.length]);

  // TTS for newly-arrived messages from the *other* couple.
  useEffect(() => {
    if (!ttsEnabled) return;
    for (const m of messages) {
      if (m.couple !== couple && !seenIds.current.has(m.id)) {
        seenIds.current.add(m.id);
        speak(`${m.speaker.endsWith("1") ? "Person 1" : "Person 2"} says: ${m.text}`);
      }
    }
    // Track our own messages so we don't replay them later if toggled on.
    for (const m of messages) seenIds.current.add(m.id);
  }, [messages, ttsEnabled, couple]);

  const send = (value: string) => {
    const v = value.trim();
    if (!v || disabled) return;
    onSend(speakerCode, v);
    setText("");
  };

  const startListening = () => {
    const Ctor = getSpeechRecognitionCtor();
    if (!Ctor) {
      alert("Speech recognition not supported in this browser.");
      return;
    }
    const rec = new Ctor();
    rec.continuous = false;
    rec.interimResults = true;
    rec.lang = "en-US";
    let final = "";
    rec.onresult = (e) => {
      final = Array.from({ length: e.results.length }, (_, i) => e.results[i][0].transcript).join("");
      setInterim(final);
    };
    rec.onend = () => {
      setListening(false);
      setInterim("");
      if (final.trim()) send(final);
    };
    rec.onerror = () => {
      setListening(false);
      setInterim("");
    };
    recRef.current = rec;
    setListening(true);
    setInterim("");
    rec.start();
  };

  const stopListening = () => {
    recRef.current?.stop();
  };

  return (
    <div style={styles.column}>
      <div style={styles.columnHeader}>
        <span style={styles.columnTitle}>{coupleLabel}</span>
        <span style={{ ...styles.speakerPick }}>
          <button
            type="button"
            style={speaker === "1" ? styles.speakerBtnActive : styles.speakerBtn}
            onClick={() => setSpeaker("1")}
          >
            Person 1
          </button>
          <button
            type="button"
            style={speaker === "2" ? styles.speakerBtnActive : styles.speakerBtn}
            onClick={() => setSpeaker("2")}
          >
            Person 2
          </button>
        </span>
      </div>

      <div ref={scrollRef} style={styles.messageList}>
        {messages.length === 0 && (
          <div style={styles.emptyMsg}>No messages yet. Say hello to start the interview.</div>
        )}
        {messages.map((m) => {
          const isP1 = m.speaker.endsWith("1");
          const isOwn = m.couple === couple;
          return (
            <div key={m.id} style={{ ...styles.bubbleRow, justifyContent: isOwn ? "flex-end" : "flex-start" }}>
              <div style={{ ...styles.bubble, ...(isOwn ? styles.bubbleOwn : styles.bubbleOther) }}>
                <div style={styles.bubbleSpeaker}>{isP1 ? "Person 1" : "Person 2"}</div>
                <div style={styles.bubbleText}>{m.text}</div>
                <div style={styles.bubbleTime}>{relativeTime(m.created_at)}</div>
              </div>
            </div>
          );
        })}
        {listening && interim && (
          <div style={{ ...styles.bubbleRow, justifyContent: "flex-end" }}>
            <div style={{ ...styles.bubble, ...styles.bubbleOwn, ...styles.bubbleInterim }}>
              <span style={styles.recDot} /> {interim}
            </div>
          </div>
        )}
      </div>

      <div style={styles.inputRow}>
        <button
          type="button"
          style={{ ...styles.micBtn, ...(listening ? styles.micBtnActive : {}) }}
          onClick={listening ? stopListening : startListening}
          title={listening ? "Stop recording" : "Start voice input"}
          disabled={disabled}
        >
          {listening ? <span style={styles.recDot} /> : "🎤"}
        </button>
        <input
          style={styles.textInput}
          placeholder={listening ? "Listening…" : `Message as ${personLabel}…`}
          value={listening ? interim || text : text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") send(text);
          }}
          disabled={disabled}
        />
        <button type="button" style={styles.sendBtn} onClick={() => send(text)} disabled={disabled || !text.trim()}>
          Send
        </button>
      </div>
    </div>
  );
}

interface ExtractionCardProps {
  extraction: InterviewExtraction;
}

function ExtractionCard({ extraction }: ExtractionCardProps) {
  const meta = AGENT_META[extraction.agent_type] ?? { icon: "🔍", label: extraction.agent_type, color: "#8b949e" };
  const pct = Math.round(extraction.confidence * 100);
  const findings = Object.entries(extraction.findings);
  return (
    <div style={{ ...styles.agentCard, borderLeftColor: meta.color }}>
      <div style={styles.agentCardHead}>
        <span style={styles.agentIcon}>{meta.icon}</span>
        <span style={styles.agentLabel}>{meta.label}</span>
        <span style={{ ...styles.agentPct, color: meta.color }}>{pct}%</span>
      </div>
      <div style={styles.confBar}>
        <div style={{ ...styles.confFill, width: `${pct}%`, background: meta.color }} />
      </div>
      {extraction.summary && <div style={styles.agentSummary}>{extraction.summary}</div>}
      {findings.length > 0 && (
        <dl style={styles.findings}>
          {findings.map(([k, v]) => (
            <div key={k} style={styles.findingRow}>
              <dt style={styles.findingKey}>{k}</dt>
              <dd style={styles.findingVal}>{String(v)}</dd>
            </div>
          ))}
        </dl>
      )}
    </div>
  );
}

export function InterviewView() {
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [labelA, setLabelA] = useState("Couple A");
  const [labelB, setLabelB] = useState("Couple B");
  const [ttsEnabled, setTtsEnabled] = useState(false);
  const [autoExtract, setAutoExtract] = useState(true);

  const { data: detail, isLoading } = useInterviewSession(sessionId);
  const createSession = useCreateInterviewSession();
  const endSession = useEndInterviewSession();
  const addMessage = useAddInterviewMessage(sessionId ?? "");
  const runExtraction = useRunExtraction(sessionId ?? "");

  const session = detail?.session;
  const messages = detail?.messages ?? [];
  const extractions = detail?.extractions ?? [];

  const messagesA = useMemo(() => messages.filter((m) => m.couple === "A"), [messages]);
  const messagesB = useMemo(() => messages.filter((m) => m.couple === "B"), [messages]);

  // Auto-run extraction every 5s while active (only if there are messages).
  const lastExtract = useRef(0);
  useEffect(() => {
    if (!sessionId || !autoExtract) return;
    if (session?.status === "completed") return;
    if (messages.length === 0) return;
    const i = setInterval(() => {
      if (Date.now() - lastExtract.current >= 5_000 && !runExtraction.isPending) {
        lastExtract.current = Date.now();
        runExtraction.mutate(undefined, { onError: () => {} });
      }
    }, 5_000);
    return () => clearInterval(i);
  }, [sessionId, autoExtract, session?.status, messages.length, runExtraction]);

  const handleCreate = () => {
    if (!labelA.trim() || !labelB.trim()) return;
    createSession.mutate(
      { couple_a_label: labelA.trim(), couple_b_label: labelB.trim() },
      {
        onSuccess: (s) => setSessionId(s.id),
      },
    );
  };

  const handleEnd = () => {
    if (!sessionId) return;
    endSession.mutate(sessionId);
  };

  const handleSend = (couple: "A" | "B") => (speaker: string, text: string) => {
    addMessage.mutate({ speaker, couple, text });
  };

  // ---- Setup screen ----
  if (!sessionId) {
    return (
      <div style={styles.view}>
        <style>{`@keyframes iv-pulse{0%,100%{opacity:1}50%{opacity:0.3}}`}</style>
        <div style={styles.setupCard}>
          <h2 style={styles.setupTitle}>Couple Interview Session</h2>
          <p style={styles.setupSub}>
            Start a live interview between two couples. AI agents extract relationship stage, wedding timeline, vendor
            interests, location, and budget in real time.
          </p>
          <label style={styles.fieldLabel}>Couple A label</label>
          <input style={styles.fieldInput} value={labelA} onChange={(e) => setLabelA(e.target.value)} autoFocus />
          <label style={styles.fieldLabel}>Couple B label</label>
          <input style={styles.fieldInput} value={labelB} onChange={(e) => setLabelB(e.target.value)} />
          <button
            type="button"
            style={styles.startBtn}
            onClick={handleCreate}
            disabled={createSession.isPending || !labelA.trim() || !labelB.trim()}
          >
            {createSession.isPending ? "Starting…" : "Start Session"}
          </button>
          {createSession.error && <div style={styles.errorMsg}>{String(createSession.error.message)}</div>}
        </div>
      </div>
    );
  }

  // ---- Active session ----
  const active = session?.status !== "completed";
  return (
    <div style={styles.view}>
      <style>{`@keyframes iv-pulse{0%,100%{opacity:1}50%{opacity:0.3}}`}</style>
      <div style={styles.topBar}>
        <div style={styles.topBarLeft}>
          <h2 style={styles.topTitle}>Couple Interview</h2>
          <span style={{ ...styles.statusPill, ...(active ? styles.statusActive : styles.statusDone) }}>
            <span style={styles.statusDot} /> {session?.status ?? "…"}
          </span>
        </div>
        <div style={styles.topBarRight}>
          <label style={styles.toggle}>
            <input type="checkbox" checked={ttsEnabled} onChange={(e) => setTtsEnabled(e.target.checked)} />
            🔊 Voice playback
          </label>
          <label style={styles.toggle}>
            <input type="checkbox" checked={autoExtract} onChange={(e) => setAutoExtract(e.target.checked)} />
            ⚡ Auto-extract
          </label>
          <button
            type="button"
            style={styles.extractBtn}
            onClick={() => {
              lastExtract.current = Date.now();
              runExtraction.mutate();
            }}
            disabled={!active || runExtraction.isPending}
          >
            {runExtraction.isPending ? "Extracting…" : "Run Extraction"}
          </button>
          <button type="button" style={styles.endBtn} onClick={handleEnd} disabled={!active || endSession.isPending}>
            {endSession.isPending ? "Ending…" : "End Session"}
          </button>
        </div>
      </div>

      {isLoading && !session && <div style={styles.emptyMsg}>Loading session…</div>}

      <div style={styles.split}>
        <ChatColumn
          coupleLabel={session?.couple_a_label ?? "Couple A"}
          couple="A"
          messages={messagesA}
          onSend={handleSend("A")}
          disabled={!active}
          ttsEnabled={ttsEnabled}
        />
        <ChatColumn
          coupleLabel={session?.couple_b_label ?? "Couple B"}
          couple="B"
          messages={messagesB}
          onSend={handleSend("B")}
          disabled={!active}
          ttsEnabled={ttsEnabled}
        />
        <div style={styles.extractionPanel}>
          <div style={styles.extractionHead}>
            <span>AI Extraction</span>
            <span style={styles.extractionCount}>{extractions.length}</span>
          </div>
          {extractions.length === 0 && (
            <div style={styles.emptyMsg}>No extractions yet. Send a few messages, then run extraction.</div>
          )}
          <div style={styles.agentList}>
            {extractions
              .slice()
              .sort((a, b) => b.confidence - a.confidence)
              .map((ex) => (
                <ExtractionCard key={ex.id} extraction={ex} />
              ))}
          </div>
        </div>
      </div>
    </div>
  );
}

// --- Self-contained styles using the app's CSS variables (dark-theme aware) ---
const styles: Record<string, React.CSSProperties> = {
  view: {
    display: "flex",
    flexDirection: "column",
    height: "100%",
    minHeight: 0,
    padding: "20px 24px",
    gap: 16,
  },
  setupCard: {
    maxWidth: 460,
    margin: "48px auto",
    background: "var(--surface)",
    border: "1px solid var(--border)",
    borderRadius: "var(--radius-lg)",
    padding: 32,
    display: "flex",
    flexDirection: "column",
    gap: 10,
  },
  setupTitle: { fontSize: 22, marginBottom: 4 },
  setupSub: { color: "var(--ink-dim)", fontSize: 14, marginBottom: 12, lineHeight: 1.5 },
  fieldLabel: { fontSize: 13, color: "var(--ink-dim)", marginTop: 6 },
  fieldInput: {
    padding: "10px 12px",
    border: "1px solid var(--border-strong)",
    borderRadius: "var(--radius-sm)",
    background: "var(--bg)",
    color: "var(--ink)",
    fontSize: 15,
  },
  startBtn: {
    marginTop: 16,
    padding: "12px 16px",
    border: "none",
    borderRadius: "var(--radius-sm)",
    background: "var(--cove-deep)",
    color: "var(--surface)",
    fontSize: 15,
    fontWeight: 600,
    cursor: "pointer",
  },
  errorMsg: { color: "var(--red)", fontSize: 13, marginTop: 8 },

  topBar: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    flexWrap: "wrap",
    gap: 12,
  },
  topBarLeft: { display: "flex", alignItems: "center", gap: 12 },
  topTitle: { fontSize: 20 },
  statusPill: {
    display: "inline-flex",
    alignItems: "center",
    gap: 6,
    padding: "4px 10px",
    borderRadius: 999,
    fontSize: 12,
    fontWeight: 600,
  },
  statusActive: { background: "color-mix(in srgb, var(--green) 22%, transparent)", color: "var(--green)" },
  statusDone: { background: "color-mix(in srgb, var(--ink-dim) 22%, transparent)", color: "var(--ink-dim)" },
  statusDot: { width: 8, height: 8, borderRadius: "50%", background: "currentColor" },
  topBarRight: { display: "flex", alignItems: "center", gap: 14, flexWrap: "wrap" },
  toggle: { display: "inline-flex", alignItems: "center", gap: 6, fontSize: 13, color: "var(--ink-dim)", cursor: "pointer" },
  extractBtn: {
    padding: "8px 14px",
    border: "1px solid var(--cove-deep)",
    borderRadius: "var(--radius-sm)",
    background: "color-mix(in srgb, var(--cove) 40%, var(--surface))",
    color: "var(--cove-deep)",
    fontSize: 13,
    fontWeight: 600,
    cursor: "pointer",
  },
  endBtn: {
    padding: "8px 14px",
    border: "1px solid var(--red)",
    borderRadius: "var(--radius-sm)",
    background: "color-mix(in srgb, var(--red) 14%, var(--surface))",
    color: "var(--red)",
    fontSize: 13,
    fontWeight: 600,
    cursor: "pointer",
  },

  split: {
    display: "grid",
    gridTemplateColumns: "1fr 1fr 1.1fr",
    gap: 14,
    flex: 1,
    minHeight: 0,
  },
  column: {
    display: "flex",
    flexDirection: "column",
    background: "var(--surface)",
    border: "1px solid var(--border)",
    borderRadius: "var(--radius-lg)",
    overflow: "hidden",
    minHeight: 0,
  },
  columnHeader: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    padding: "12px 14px",
    borderBottom: "1px solid var(--border)",
    background: "var(--bg-alt)",
  },
  columnTitle: { fontWeight: 600, fontSize: 15 },
  speakerPick: { display: "inline-flex", gap: 4 },
  speakerBtn: {
    padding: "3px 8px",
    border: "1px solid var(--border-strong)",
    borderRadius: 999,
    background: "transparent",
    color: "var(--ink-dim)",
    fontSize: 11,
    cursor: "pointer",
  },
  speakerBtnActive: {
    padding: "3px 8px",
    border: "1px solid var(--cove-deep)",
    borderRadius: 999,
    background: "var(--cove-deep)",
    color: "var(--surface)",
    fontSize: 11,
    fontWeight: 600,
    cursor: "pointer",
  },

  messageList: {
    flex: 1,
    overflowY: "auto",
    padding: "12px 14px",
    display: "flex",
    flexDirection: "column",
    gap: 8,
    minHeight: 0,
  },
  emptyMsg: { color: "var(--ink-dim)", fontSize: 13, textAlign: "center", padding: "24px 8px" },
  bubbleRow: { display: "flex", width: "100%" },
  bubble: {
    maxWidth: "82%",
    padding: "8px 12px",
    borderRadius: "var(--radius)",
    fontSize: 14,
    lineHeight: 1.4,
  },
  bubbleOwn: { background: "var(--cove-deep)", color: "var(--surface)" },
  bubbleOther: { background: "var(--chip)", color: "var(--ink)" },
  bubbleInterim: { opacity: 0.8, fontStyle: "italic" },
  bubbleSpeaker: { fontSize: 11, opacity: 0.8, marginBottom: 2 },
  bubbleText: { wordBreak: "break-word" },
  bubbleTime: { fontSize: 10, opacity: 0.6, marginTop: 4, textAlign: "right" },

  inputRow: {
    display: "flex",
    gap: 8,
    padding: "10px 12px",
    borderTop: "1px solid var(--border)",
    background: "var(--bg-alt)",
  },
  micBtn: {
    width: 38,
    height: 38,
    flexShrink: 0,
    border: "1px solid var(--border-strong)",
    borderRadius: "var(--radius-sm)",
    background: "var(--surface)",
    color: "var(--ink)",
    fontSize: 18,
    cursor: "pointer",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
  },
  micBtnActive: { borderColor: "var(--red)", background: "color-mix(in srgb, var(--red) 16%, var(--surface))" },
  recDot: {
    display: "inline-block",
    width: 10,
    height: 10,
    borderRadius: "50%",
    background: "var(--red)",
    animation: "iv-pulse 1s infinite",
  },
  textInput: {
    flex: 1,
    padding: "9px 12px",
    border: "1px solid var(--border-strong)",
    borderRadius: "var(--radius-sm)",
    background: "var(--surface)",
    color: "var(--ink)",
    fontSize: 14,
  },
  sendBtn: {
    padding: "0 16px",
    border: "none",
    borderRadius: "var(--radius-sm)",
    background: "var(--cove-deep)",
    color: "var(--surface)",
    fontSize: 14,
    fontWeight: 600,
    cursor: "pointer",
  },

  extractionPanel: {
    display: "flex",
    flexDirection: "column",
    background: "var(--surface)",
    border: "1px solid var(--border)",
    borderRadius: "var(--radius-lg)",
    overflow: "hidden",
    minHeight: 0,
  },
  extractionHead: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    padding: "12px 14px",
    borderBottom: "1px solid var(--border)",
    background: "var(--bg-alt)",
    fontWeight: 600,
    fontSize: 15,
  },
  extractionCount: {
    background: "var(--chip)",
    color: "var(--ink-dim)",
    borderRadius: 999,
    padding: "1px 8px",
    fontSize: 12,
  },
  agentList: { flex: 1, overflowY: "auto", padding: 12, display: "flex", flexDirection: "column", gap: 10, minHeight: 0 },
  agentCard: {
    background: "var(--bg)",
    border: "1px solid var(--border)",
    borderLeft: "4px solid var(--ink-dim)",
    borderRadius: "var(--radius)",
    padding: "10px 12px",
  },
  agentCardHead: { display: "flex", alignItems: "center", gap: 8, marginBottom: 6 },
  agentIcon: { fontSize: 16 },
  agentLabel: { fontWeight: 600, fontSize: 13, flex: 1 },
  agentPct: { fontSize: 13, fontWeight: 700 },
  confBar: { height: 5, borderRadius: 999, background: "var(--ink-faint)", overflow: "hidden", marginBottom: 8 },
  confFill: { height: "100%", borderRadius: 999, transition: "width 0.4s ease" },
  agentSummary: { fontSize: 13, color: "var(--ink)", lineHeight: 1.45, marginBottom: 6 },
  findings: { margin: 0, display: "flex", flexDirection: "column", gap: 4 },
  findingRow: { display: "flex", gap: 8, fontSize: 12 },
  findingKey: { color: "var(--ink-dim)", minWidth: 90, flexShrink: 0 },
  findingVal: { margin: 0, color: "var(--ink)", wordBreak: "break-word" },
};
