export function ViewSkeleton() {
  return (
    <main className="z-container z-page z-skeleton-page" aria-busy="true" aria-label="Loading">
      <section className="z-tile z-skeleton-card">
        <span className="z-skeleton-line w-32" />
        <span className="z-skeleton-line w-64" />
        <span className="z-skeleton-line w-48" />
        <span className="z-skeleton-line w-56" />
      </section>
    </main>
  );
}
