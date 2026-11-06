import { Component, type ErrorInfo, type ReactNode } from "react";
import { Button } from "@cloudflare/kumo";

interface Props { children: ReactNode }
interface State { failed: boolean }

export class ErrorBoundary extends Component<Props, State> {
  state: State = { failed: false };

  static getDerivedStateFromError(): State {
    return { failed: true };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("ui error", error, info);
  }

  render() {
    if (!this.state.failed) return this.props.children;
    return (
      <main className="z-container z-page z-error-page">
        <section className="z-tile z-error-card" role="alert">
          <h1 className="z-page-title">مشکلی پیش آمد</h1>
          <p className="z-error-text">یک خطای غیرمنتظره رخ داد. صفحه را دوباره بارگذاری کن.</p>
          <div className="z-error-actions">
            <Button variant="primary" onClick={() => window.location.reload()}>
              بارگذاری مجدد
            </Button>
          </div>
        </section>
      </main>
    );
  }
}
