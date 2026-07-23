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
        "html": "BSc in Pure Mathematics &amp; Computer Science, Maynooth University, completed May 2026, with a year abroad at the University of Ottawa. Run a two-node k3s cluster on a single Proxmox host: hardware and virtualisation up through GitOps-managed Kubernetes, with CI/CD, monitoring, and alerting. Troubleshoot faults across the stack, from hardware quirks and kernel drivers up to Kubernetes workloads. Targeting platform engineering, site reliability, and low-level infrastructure."
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
            "items": "TCP/IP, DNS (split-horizon), WireGuard, Cloudflare Tunnel + Access, SSL/TLS"
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
          "Cluster-wide hardening: NetworkPolicy default-deny per namespace, non-root SecurityContext with dropped capabilities and read-only root filesystem.",
          "CI/CD end-to-end: GitHub Actions builds, pushes to GHCR, updates SHA-pinned manifest, ArgoCD reconciles; in-cluster runner lints manifests.",
          "Public ingress via Cloudflare Tunnel (no inbound ports open), split-horizon DNS via Technitium, wildcard TLS via cert-manager and Let’s Encrypt DNS-01; Cloudflare Access on sensitive services.",
          "Observability with VictoriaMetrics, Grafana, Alertmanager and <code>absent()</code> guards so a dead scrape self-alerts; encrypted offsite restic backups to Backblaze B2, quarterly verified restores.",
          "Diagnosed intermittent e1000e (Intel I219-LM) NIC transmit hangs across ethtool offload settings, modprobe driver parameters, and GRUB <code>pcie_aspm=off</code>.",
          "Architecture decisions recorded as ADRs, incidents as post-mortems. <a href=\"https://github.com/hpdowd/homelab\" target=\"_blank\" rel=\"noopener\">github.com/hpdowd/homelab</a>"
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
      },
      {
        "kind": "entry",
        "title": "PDF-Sorter — Workplace Automation Tool",
        "org": "Python",
        "dates": "2025",
        "bullets": [
          "Regex-based tool to auto-sort and rename PDFs by filename and content metadata, with OCR fallback; deployed during HR placement at Roscommon County Council. <a href=\"https://github.com/hpdowd/PDF-Sorter\" target=\"_blank\" rel=\"noopener\">github.com/hpdowd/PDF-Sorter</a>"
        ]
      }
    ]
  },
  {
    "heading": "Experience",
    "blocks": [
      {
        "kind": "entry",
        "title": "Summer Student — Human Resources",
        "org": "Roscommon County Council",
        "dates": "Jun – Sep 2025",
        "bullets": [
          "Data entry, records management, and financial processing across employee files and payroll administration.",
          "Identified a repetitive document-sorting bottleneck and shipped the PDF-Sorter tool to automate it."
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
        "subtitle": "Pure Mathematics & Computer Science"
      },
      {
        "kind": "prose",
        "html": "<strong>Relevant modules:</strong> Operating Systems, Communications &amp; Concurrency; Algorithms &amp; Data Structures; Software Verification; Programming Languages &amp; Compilers; Machine Learning &amp; Neural Networks."
      },
      {
        "kind": "entry",
        "title": "Study Abroad",
        "org": "University of Ottawa",
        "dates": "Sep 2024 – May 2025",
        "bullets": []
      },
      {
        "kind": "prose",
        "html": "Fortinet NSE 1–3 (2026) · Full driving licence · CTYI alumnus"
      }
    ]
  }
];
