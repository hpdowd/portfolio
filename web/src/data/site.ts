/**
 * Single source of truth for identity, contact links, and structured data.
 *
 * The Person/WebSite JSON-LD used to be declared twice (index + cv) and had
 * already drifted (one carried `jobTitle`, the other `description`/`address`).
 * Everything identity-shaped now lives here so the pages can't diverge; the
 * builders below emit the exact schema.org nodes each page embeds.
 */

const SITE_URL = 'https://henrydowd.dev';
const PERSON_ID = `${SITE_URL}/#person`;
const WEBSITE_ID = `${SITE_URL}/#website`;

export const person = {
  name: 'Henry Dowd',
  url: SITE_URL,
  email: 'henry@dowd.ie',
  jobTitle: 'Infrastructure engineer',
  description:
    'BSc in Pure Mathematics & Computer Science from Maynooth University, targeting graduate infrastructure engineering / systems / SRE roles.',
  alumniOf: 'Maynooth University',
  addressCountry: 'IE',
  github: 'https://github.com/hpdowd',
  linkedin: 'https://www.linkedin.com/in/henrypdowd/',
  knowsAbout: [
    'Infrastructure Engineering',
    'Linux',
    'Computer Networking',
    'Virtualization',
    'Kubernetes',
    'GitOps',
    'Site Reliability Engineering',
    'Observability',
  ],
} as const;

/** External profiles + email, in the order the masthead lists them. */
export const social: { key: string; href: string; value: string; external?: boolean }[] = [
  { key: 'Email', href: `mailto:${person.email}`, value: person.email },
  { key: 'GitHub', href: person.github, value: 'github.com/hpdowd', external: true },
  { key: 'LinkedIn', href: person.linkedin, value: 'in/henrypdowd', external: true },
];

/** The schema.org Person node, shared by the home @graph and the standalone CV. */
export function personNode() {
  return {
    '@type': 'Person',
    '@id': PERSON_ID,
    name: person.name,
    url: person.url,
    email: person.email,
    jobTitle: person.jobTitle,
    description: person.description,
    address: { '@type': 'PostalAddress', addressCountry: person.addressCountry },
    alumniOf: { '@type': 'CollegeOrUniversity', name: person.alumniOf },
    knowsAbout: [...person.knowsAbout],
    sameAs: [person.github, person.linkedin],
  };
}

/** Home page @graph: the site plus its author. */
export function homeGraphLd() {
  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'WebSite',
        '@id': WEBSITE_ID,
        url: person.url,
        name: person.name,
        description:
          'Infrastructure-engineering portfolio, self-hosted on a Kubernetes homelab and reporting live telemetry from it.',
        inLanguage: 'en',
        author: { '@id': PERSON_ID },
      },
      personNode(),
    ],
  };
}

/** Standalone Person node for the CV page (same data, own @context). */
export function personLd() {
  return { '@context': 'https://schema.org', ...personNode() };
}
