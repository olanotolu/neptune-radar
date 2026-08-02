import { Component, type ErrorInfo, type ReactNode } from "react";

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

// ErrorBoundary catches render-time errors in the component tree and shows
// a recovery screen instead of a blank white page. The dashboard has 18+
// views with live data; one bad payload shouldn't kill the whole app.
export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("[ErrorBoundary]", error, info.componentStack);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="error-boundary">
          <div className="error-boundary__card">
            <h2 className="error-boundary__title">Something went wrong</h2>
            <p className="error-boundary__msg">
              {this.state.error?.message || "An unexpected error occurred."}
            </p>
            <button
              className="error-boundary__btn"
              onClick={() => {
                this.setState({ hasError: false, error: undefined });
                window.location.hash = "#/today";
              }}
            >
              Back to Today
            </button>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
