// AUTO-GENERATED from web/cv/resume.tex by `npm run cv:data`.
// Do not edit by hand — edit the .tex and regenerate.

export interface SkillRow { label: string; items: string; }
export interface CvEntry { kind: 'entry'; title: string; org: string; dates: string; subtitle?: string; bullets: string[]; }
export interface CvSkills { kind: 'skills'; rows: SkillRow[]; }
export interface CvProse { kind: 'prose'; html: string; }
export type CvBlock = CvEntry | CvSkills | CvProse;
export interface CvSection { heading: string; blocks: CvBlock[]; }

export const sections: CvSection[] = [
  {
    "heading": "Profile",
    "blocks": [
      {
        "kind": "prose",
        "html": "Production Engineer on the Fleet Operations team at Crusoe, working on the Managed Kubernetes and Managed VM platforms. Run a two-node k3s cluster on a single Proxmox host at home, from the hardware and virtualisation up through GitOps-managed Kubernetes, and troubleshoot faults across that stack from kernel drivers to workloads."
      }
    ]
  },
  {
    "heading": "Technical Skills",
    "blocks": [
      {
        "kind": "skills",
        "rows": [
          {
            "label": "Linux & Virtualisation",
            "items": "Arch Linux, Proxmox VE, LXC, ZFS, systemd, shell scripting"
          },
          {
            "label": "Networking & Security",
            "items": "TCP/IP, DNS (split-horizon), WireGuard, Cloudflare Tunnel + Access, SSL/TLS, Authelia (OIDC, ForwardAuth)"
          },
          {
            "label": "Kubernetes & GitOps",
            "items": "k3s, Helm, ArgoCD (app-of-apps), MetalLB, Traefik, Longhorn, Sealed Secrets, cert-manager"
          },
          {
            "label": "CI/CD & Observability",
            "items": "GitHub Actions, GHCR, VictoriaMetrics, Grafana, Alertmanager, restic, Backblaze B2"
          },
          {
            "label": "Languages & Tools",
            "items": "Python, Bash, Java, C, SQL; Git, Gitea, Technitium DNS"
          }
        ]
      }
    ]
  },
  {
    "heading": "Experience",
    "blocks": [
      {
        "kind": "entry",
        "title": "Production Engineer",
        "org": "Crusoe",
        "dates": "Aug 2026 – Present",
        "bullets": [
          "Fleet Operations for Crusoe’s Managed Kubernetes and Managed VM platforms, run for external customers.",
          "Kubernetes platform tooling, monitoring and alerting, log analysis, incident response and post-mortems, automated remediation, and SLI/SLO maintenance."
        ]
      },
      {
        "kind": "entry",
        "title": "Summer Student — Human Resources",
        "org": "Roscommon County Council",
        "dates": "Jun – Sep 2025",
        "bullets": [
          "Data entry, records management, and financial processing across employee files and payroll administration; shipped PDF-Sorter (Python, regex + OCR fallback) to automate a document-sorting bottleneck. <a href=\"https://github.com/hpdowd/PDF-Sorter\" target=\"_blank\" rel=\"noopener\">github.com/hpdowd/PDF-Sorter</a>"
        ]
      }
    ]
  },
  {
    "heading": "Projects",
    "blocks": [
      {
        "kind": "entry",
        "title": "Homelab — k3s / GitOps Platform",
        "org": "Personal",
        "dates": "2024 – Present",
        "bullets": [
          "Two-node k3s cluster on a single Proxmox host, GitOps via ArgoCD app-of-apps pattern, git repo as single source of truth.",
          "Platform assembled from components: MetalLB, Traefik ingress, Longhorn persistent block storage (replica=1, deliberate single-worker tradeoff over host-level ZFS RAID-1 mirror), Sealed Secrets for encrypted credentials in git.",
          "Workload hardening: dropped capabilities, seccomp RuntimeDefault and no privilege escalation across the app workloads; default-deny ingress NetworkPolicies on the namespaces that tolerate them.",
          "CI/CD end-to-end: GitHub Actions builds, pushes to GHCR, updates SHA-pinned manifest, ArgoCD reconciles; in-cluster runner lints manifests.",
          "Public ingress via Cloudflare Tunnel (no inbound ports open), split-horizon DNS via Technitium, wildcard TLS via cert-manager and Let’s Encrypt DNS-01; Cloudflare Access on sensitive hosts.",
          "Single sign-on with Authelia: Traefik ForwardAuth on the gated hosts, OIDC provider for six applications, LAN paths ungated as break-glass.",
          "Observability with VictoriaMetrics, Grafana, Alertmanager and <code>absent()</code> guards so a dead scrape self-alerts; encrypted offsite restic backups to Backblaze B2, quarterly verified restores.",
          "Diagnosed intermittent e1000e (Intel I219-LM) NIC transmit hangs across ethtool offload settings, modprobe driver parameters, and GRUB <code>pcie_aspm=off</code>; architecture decisions recorded as ADRs, incidents as post-mortems. <a href=\"https://github.com/hpdowd/homelab\" target=\"_blank\" rel=\"noopener\">github.com/hpdowd/homelab</a>"
        ]
      },
      {
        "kind": "entry",
        "title": "Final Year Project — Paraphrase Detection",
        "org": "Maynooth University",
        "dates": "2025 – 2026",
        "bullets": [
          "Built a dual-branch classifier fusing BERT cosine similarity, Zhang–Shasha tree edit distance, Jaccard overlap, and length ratio; evaluated on the MSRP corpus in Python with HuggingFace Transformers, PyTorch, and scikit-learn.",
          "Performed error analysis comparing structural vs. semantic methods, with confusion matrices, score distributions, and parse-tree examples of failure modes. <a href=\"https://github.com/hpdowd/paraphrase_detector\" target=\"_blank\" rel=\"noopener\">github.com/hpdowd/paraphrase_detector</a>"
        ]
      }
    ]
  },
  {
    "heading": "Education & Certifications",
    "blocks": [
      {
        "kind": "entry",
        "title": "BSc Computational Thinking",
        "org": "Maynooth University",
        "dates": "Sep 2022 – May 2026",
        "bullets": [],
        "subtitle": "Pure Mathematics & Computer Science · Year abroad at the University of Ottawa, Sep 2024 – May 2025"
      },
      {
        "kind": "prose",
        "html": "<strong>Relevant modules:</strong> Operating Systems, Communications &amp; Concurrency; Algorithms &amp; Data Structures; Software Verification; Programming Languages &amp; Compilers; Machine Learning &amp; Neural Networks.</p><p>Fortinet NSE 1–3 (2026) · Full driving licence · CTYI alumnus"
      }
    ]
  }
];
