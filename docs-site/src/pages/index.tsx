import Link from "@docusaurus/Link";
import Layout from "@theme/Layout";

export default function Home(): JSX.Element {
  return (
    <Layout title="Zero Trust Control Plane" description="Documentation">
      <main style={{ padding: "2rem", textAlign: "center" }}>
        <h1>Zero Trust Control Plane</h1>
        <p>Documentation for the zero-trust control plane backend and clients.</p>
        <Link to="/docs/intro" className="button button--primary button--lg">
          Get started
        </Link>
      </main>
    </Layout>
  );
}
