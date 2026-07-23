// AUTO-GENERATED from web/cv/primary_cv.tex by `npm run cv:data`.
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
        "html": "BSc in Pure Mathematics &amp; Computer Science, Maynooth University, completed May 2026, with a year abroad at the University of Ottawa. Run a two-node k3s cluster on a single Proxmox host, administering the stack from physical hardware and virtualisation up to GitOps-managed Kubernetes, with CI/CD, monitoring, and alerting built in. Comfortable with Linux administration, networking, and Python scripting, and troubleshoot my own infrastructure when it breaks. Targeting roles in platform engineering, site reliability, DevOps, and low-level infrastructure."
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
            "items": "Arch Linux, Proxmox VE, LXC, ZFS, systemd, shell scripting, SSH key management"
          },
          {
            "label": "Networking",
            "items": "TCP/IP, DNS (split-horizon), DHCP, WireGuard, Cloudflare Tunnels, reverse proxies / ingress, SSL/TLS"
          },
          {
            "label": "Containers & CI/CD",
            "items": "Kubernetes (k3s), Docker, Helm, ArgoCD (GitOps), GitHub Actions, MetalLB, Traefik, Longhorn"
          },
          {
            "label": "Observability & Backup",
            "items": "VictoriaMetrics, Grafana, Alertmanager, restic, Backblaze B2"
          },
          {
            "label": "Scripting & Tools",
            "items": "Python, Bash, Java, C, SQL; Git, Gitea, cert-manager (Let’s Encrypt), Technitium DNS, LaTeX"
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
        "title": "Self-Hosted Infrastructure &amp; Home Lab",
        "org": "Personal",
        "dates": "2024 – Present",
        "bullets": [
          "Two-node k3s cluster on single Proxmox host, GitOps via ArgoCD app-of-apps pattern, git repo as single source of truth.",
          "Platform assembled from components: MetalLB load balancing, Traefik ingress, Longhorn persistent block storage (replica=1, deliberate single-worker tradeoff), Sealed Secrets for encrypted credentials in git.",
          "CI/CD end-to-end: GitHub Actions builds, pushes to GHCR, updates SHA-pinned manifest, ArgoCD reconciles; in-cluster runner lints manifests.",
          "Public ingress via Cloudflare Tunnel (no inbound ports open), split-horizon DNS via Technitium, wildcard TLS via cert-manager and Let’s Encrypt DNS-01.",
          "Observability stack VictoriaMetrics, Grafana, Alertmanager with <code>absent()</code> guards so a dead scrape self-alerts; encrypted offsite restic backups to Backblaze B2, quarterly verified restores.",
          "Diagnosed intermittent e1000e (Intel I219-LM) NIC transmit hangs across ethtool offload settings, modprobe driver parameters, and GRUB <code>pcie_aspm=off</code> — layered isolation separating configuration from firmware.",
          "Decisions documented as ADRs, incidents as post-mortems. <a href=\"https://github.com/hpdowd/homelab\" target=\"_blank\" rel=\"noopener\">github.com/hpdowd/homelab</a>"
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
          "Data entry, records management, and financial processing across employee files and payroll-adjacent administration.",
          "Identified a repetitive document-sorting bottleneck and built a custom Python tool to automate the workflow."
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
        "html": "Fortinet NSE 1–3 (2026) · Full driving licence"
      }
    ]
  }
];
