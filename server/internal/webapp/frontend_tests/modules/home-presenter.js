(function (root, factory) {
  const api = factory();
  if (typeof module === "object" && module.exports) module.exports = api;
  root.VyrdenHome = api;
})(typeof globalThis !== "undefined" ? globalThis : window, function () {
  const RELEASE_TOKENS = [
    "webrip",
    "web-dl",
    "webdl",
    "bluray",
    "brrip",
    "bdrip",
    "dvdrip",
    "hdrip",
    "remux",
    "proper",
    "repack",
    "extended",
    "criterion",
    "unrated",
    "directors cut",
    "director's cut",
    "x264",
    "x265",
    "h264",
    "h265",
    "hevc",
    "av1",
    "aac",
    "ac3",
    "dts",
    "truehd",
    "atmos",
    "yify",
    "rarbg",
    "etrg",
    "amzn",
    "nf",
    "dsnp",
    "ddp5 1",
    "ddp 5 1",
    "2160p",
    "1080p",
    "720p",
    "480p",
    "hdr",
    "uhd",
    "10bit",
  ];

  const PLACEHOLDER_PALETTES = [
    { base: "#0b0f14", mid: "#171d25", edge: "#0f141b", glow: "#4ebfae", accent: "#6f8ea6", text: "#f4f1ea" },
    { base: "#0d1116", mid: "#1b2129", edge: "#111820", glow: "#58b5c4", accent: "#7b93a8", text: "#f4f1ea" },
    { base: "#0c1015", mid: "#1a1f27", edge: "#11161d", glow: "#4fa794", accent: "#8a7b6a", text: "#f4f1ea" },
    { base: "#0f1014", mid: "#1d1c22", edge: "#14161b", glow: "#758fb1", accent: "#7a8f9f", text: "#f4f1ea" },
    { base: "#0c1114", mid: "#172128", edge: "#10191f", glow: "#52baa6", accent: "#758e9c", text: "#f4f1ea" },
    { base: "#111015", mid: "#221d23", edge: "#18161a", glow: "#8d8cae", accent: "#8f8475", text: "#f4f1ea" },
    { base: "#0d1218", mid: "#1b242d", edge: "#121a22", glow: "#5eb4bf", accent: "#748e9e", text: "#f4f1ea" },
    { base: "#101216", mid: "#20252d", edge: "#171b22", glow: "#66a6b8", accent: "#8a8578", text: "#f4f1ea" },
  ];

  function escapeHTML(value) {
    return String(value ?? "").replace(/[&<>"']/g, char => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#039;" }[char]));
  }

  function escapeXML(value = "") {
    return String(value ?? "").replace(/[&<>"']/g, char => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&apos;" }[char]));
  }

  function repairText(value = "") {
    return String(value || "")
      .replace(/\u00C2/g, "")
      .replace(/\u00C3\u00A9/g, "e")
      .replace(/\u00C3\u00A8/g, "e")
      .replace(/\u00C3\u00AA/g, "e")
      .replace(/\u00C3\u00A1/g, "a")
      .replace(/\u00C3\u00A2/g, "a")
      .replace(/\u00C3\u00AD/g, "i")
      .replace(/\u00C3\u00B3/g, "o")
      .replace(/\u00C3\u00BA/g, "u")
      .replace(/\u00C3\u00B1/g, "n")
      .replace(/\u00E2\u20AC\u201C|\u00E2\u20AC\u201D/g, "-")
      .replace(/\u00E2\u20AC\u02DC|\u00E2\u20AC\u2122/g, "'")
      .replace(/\u00E2\u20AC\u0153|\u00E2\u20AC\u009D/g, "\"");
  }

  function normalizeWhitespace(value = "") {
    return repairText(String(value || "")).replace(/\s+/g, " ").trim();
  }

  function stripExtension(value = "") {
    return String(value || "").replace(/\.[a-z0-9]{2,4}$/i, "");
  }

  function looksLikeJunk(value = "") {
    const cleaned = normalizeWhitespace(String(value || ""));
    if (!cleaned) return true;
    return /^(sample|movie|episode|video|unknown)$/i.test(cleaned);
  }

  function cleanDisplayTitle(raw = "") {
    let value = stripExtension(raw)
      .replace(/[_]/g, " ")
      .replace(/\./g, " ")
      .replace(/\s+-\s+/g, " ")
      .replace(/\[[^\]]*\]/g, " ")
      .replace(/\([^)]*(webrip|bluray|x264|x265|h\.?264|h\.?265|remux|hdr|2160p|1080p|720p)[^)]*\)/gi, " ")
      .replace(/\bS\d{1,2}E\d{1,2}(?:E\d{1,2})?\b/gi, " ")
      .replace(/\b(19|20)\d{2}\b/g, " ");
    for (const token of RELEASE_TOKENS) {
      value = value.replace(new RegExp(`\\b${escapeRegex(token)}\\b`, "gi"), " ");
    }
    value = value
      .replace(/\b(group|yts|etrg|rarbg)\b$/i, " ")
      .replace(/\(\s*\)/g, " ")
      .replace(/[-:]\s*$/g, " ");
    value = normalizeWhitespace(value);
    return value || normalizeWhitespace(stripExtension(raw));
  }

  function extractYear(value = "", fallback = "") {
    const match = repairText(String(value || fallback || "")).match(/\b(19|20)\d{2}\b/);
    return match ? match[0] : "";
  }

  function extractEpisodeCode(title = "", subtitle = "") {
    const match = repairText(String(`${title} ${subtitle}`)).match(/\bS(\d{1,2})E(\d{1,2})(?:E(\d{1,2}))?\b/i);
    if (!match) return "";
    const season = Number(match[1]);
    const start = Number(match[2]);
    const end = Number(match[3] || 0);
    if (end && end !== start) return `S${season} \u00B7 E${start}-${end}`;
    return `S${season} \u00B7 E${start}`;
  }

  function cleanDisplaySubtitle(item = {}) {
    const rawSubtitle = normalizeWhitespace(item.subtitle || "");
    const episodeCode = extractEpisodeCode(item.title, rawSubtitle);
    if (episodeCode) return episodeCode;
    const year = extractYear(rawSubtitle, item.title);
    if (year) return year;
    if (/resume from /i.test(rawSubtitle)) return "";
    if (/^\d+\s+season/i.test(rawSubtitle)) return rawSubtitle.replace(/\s*\/\s*/g, " / ");
    const cleaned = cleanDisplayTitle(rawSubtitle);
    return looksLikeJunk(cleaned) ? "" : cleaned;
  }

  function kindLabel(kind = "") {
    const normalized = String(kind || "").toLowerCase();
    if (normalized === "movie" || normalized === "movies") return "Movie";
    if (normalized === "series" || normalized === "tv") return "TV";
    if (normalized === "episode") return "Episode";
    if (normalized === "empty") return "Setup";
    return normalized ? normalized[0].toUpperCase() + normalized.slice(1) : "";
  }

  function cleanDisplayMeta(item = {}, display = {}) {
    const parts = [];
    if (display.displayYear) parts.push(display.displayYear);
    if (display.displaySubtitle && display.displaySubtitle !== display.displayYear) parts.push(display.displaySubtitle);
    const label = kindLabel(item.kind);
    if (label) parts.push(label);
    return parts.join(" \u00B7 ");
  }

  function cleanDescription(value = "") {
    const text = normalizeWhitespace(value);
    if (!text) return "";
    return text.length > 220 ? `${text.slice(0, 217).trimEnd()}...` : text;
  }

  function presentHomeItem(item = {}, options = {}) {
    const titleSource = firstNonEmpty(options.title, item.title, item.name, item.relPath, "Untitled");
    const displayTitle = cleanDisplayTitle(titleSource);
    const displaySubtitle = firstNonEmpty(options.displaySubtitle, cleanDisplaySubtitle(item));
    const displayYear = firstNonEmpty(options.displayYear, extractYear(item.subtitle, titleSource));
    const posterUrl = firstNonEmpty(options.posterUrl, item.posterUrl);
    const explicitBackdropUrl = firstNonEmpty(options.backdropUrl, item.backdropUrl);
    const backdropUrl = explicitBackdropUrl || posterUrl;
    const thumbUrl = firstNonEmpty(options.thumbUrl, item.backdropUrl, item.posterUrl, backdropUrl);
    const displayDescription = cleanDescription(firstNonEmpty(options.description, item.description, item.overview));
    const display = {
      id: item.id || item.mediaSourceId || "",
      kind: item.kind || "",
      mediaSourceId: item.mediaSourceId || "",
      route: item.route || "",
      progressPercent: clampPercent(item.progressPercent ?? item.percent),
      displayTitle,
      displaySubtitle,
      displayYear,
      displayMeta: firstNonEmpty(options.displayMeta, cleanDisplayMeta(item, { displayYear, displaySubtitle })),
      displayDescription,
      posterUrl,
      backdropUrl,
      thumbUrl,
      usePosterBackdrop: !explicitBackdropUrl && Boolean(posterUrl),
      hasArtwork: Boolean(posterUrl || backdropUrl || thumbUrl),
      needsPlaceholder: !posterUrl && !backdropUrl && !thumbUrl,
      syntheticArtworkOnly: Boolean(options.syntheticArtworkOnly),
      searchText: normalizeWhitespace([displayTitle, displaySubtitle, displayYear, item.route, kindLabel(item.kind)].filter(Boolean).join(" ")),
    };
    return display;
  }

  function buildPlaceholderArtwork(item = {}, variant = "poster") {
    const title = cleanDisplayTitle(item.displayTitle || item.title || "Untitled");
    const subtitle = normalizeWhitespace(item.displaySubtitle || item.displayYear || kindLabel(item.kind) || "");
    const meta = normalizeWhitespace(item.displayMeta || subtitle || "");
    const seed = visualSeed(`${title}|${item.displayYear || ""}|${item.kind || ""}`);
    const palette = placeholderPalette(seed);
    if (variant === "hero") return svgDataUrl(buildHeroSvg(title, meta, palette, seed));
    if (variant === "wide") return svgDataUrl(buildWideSvg(title, subtitle, palette, seed));
    if (variant === "activity") return svgDataUrl(buildActivitySvg(title, subtitle, palette, seed));
    return svgDataUrl(buildPosterSvg(title, subtitle, palette, seed));
  }

  function buildPosterSvg(title, subtitle, palette, seed = 0) {
    return `<svg xmlns="http://www.w3.org/2000/svg" width="600" height="900" viewBox="0 0 600 900">
  <defs>
    <linearGradient id="bg" x1="0" x2="1" y1="0" y2="1">
      <stop offset="0%" stop-color="${palette.base}"/>
      <stop offset="52%" stop-color="${palette.mid}"/>
      <stop offset="100%" stop-color="${palette.edge}"/>
    </linearGradient>
    <radialGradient id="glow" cx="0.82" cy="0.12" r="0.72">
      <stop offset="0%" stop-color="${palette.glow}" stop-opacity="0.58"/>
      <stop offset="100%" stop-color="${palette.glow}" stop-opacity="0"/>
    </radialGradient>
    <linearGradient id="vignette" x1="0" x2="0" y1="0" y2="1">
      <stop offset="0%" stop-color="#000000" stop-opacity="0"/>
      <stop offset="72%" stop-color="#000000" stop-opacity="0.12"/>
      <stop offset="100%" stop-color="#000000" stop-opacity="0.4"/>
    </linearGradient>
  </defs>
  <rect width="600" height="900" fill="url(#bg)"/>
  <rect width="600" height="900" fill="url(#glow)"/>
  ${posterSceneMarkup(palette, seed)}
  <rect width="600" height="900" fill="url(#glow)" opacity="0.32"/>
  <rect width="600" height="900" fill="url(#vignette)"/>
  <rect x="38" y="38" width="524" height="824" rx="26" fill="none" stroke="#f4f1ea" stroke-opacity="0.1" stroke-width="2"/>
  <path d="M58 564c82-40 158-56 246-58 94-2 178 20 238 62v204H58z" fill="#06090d" fill-opacity="0.25"/>
  <rect x="58" y="558" width="98" height="4" rx="2" fill="${palette.glow}" fill-opacity="0.74"/>
  <rect x="58" y="612" width="484" height="214" rx="22" fill="#080c11" fill-opacity="0.16"/>
  <rect x="58" y="612" width="484" height="214" rx="22" fill="#f4f1ea" fill-opacity="0.02"/>
</svg>`;
  }

  function buildHeroSvg(title, meta, palette, seed = 0) {
    const monogram = escapeXML(initials(title).slice(0, 2) || "V");
    return `<svg xmlns="http://www.w3.org/2000/svg" width="1280" height="720" viewBox="0 0 1280 720">
  <defs>
    <linearGradient id="bg" x1="0" x2="1" y1="0" y2="1">
      <stop offset="0%" stop-color="${palette.base}"/>
      <stop offset="50%" stop-color="${palette.mid}"/>
      <stop offset="100%" stop-color="${palette.edge}"/>
    </linearGradient>
    <radialGradient id="glowLeft" cx="0.18" cy="0.24" r="0.42">
      <stop offset="0%" stop-color="${palette.accent}" stop-opacity="0.32"/>
      <stop offset="100%" stop-color="${palette.accent}" stop-opacity="0"/>
    </radialGradient>
    <radialGradient id="glowRight" cx="0.82" cy="0.28" r="0.48">
      <stop offset="0%" stop-color="${palette.glow}" stop-opacity="0.46"/>
      <stop offset="100%" stop-color="${palette.glow}" stop-opacity="0"/>
    </radialGradient>
    <linearGradient id="haze" x1="0" x2="0" y1="0" y2="1">
      <stop offset="0%" stop-color="#ffffff" stop-opacity="0.02"/>
      <stop offset="100%" stop-color="#000000" stop-opacity="0.34"/>
    </linearGradient>
  </defs>
  <rect width="1280" height="720" fill="url(#bg)"/>
  <rect width="1280" height="720" fill="url(#glowLeft)"/>
  <rect width="1280" height="720" fill="url(#glowRight)"/>
  ${heroSceneMarkup(palette, seed)}
  <rect width="1280" height="720" fill="url(#haze)"/>
  <path d="M0 566c212-64 394-88 590-68 214 22 398 88 690 222v0H0z" fill="#050911" fill-opacity="0.48"/>
  <rect x="80" y="102" width="350" height="232" rx="36" fill="#020611" fill-opacity="0.12"/>
  <text x="900" y="414" fill="#f5f8ff" fill-opacity="0.08" font-family="Outfit, Manrope, system-ui, sans-serif" font-size="236" font-weight="800">${monogram}</text>
  <rect x="84" y="618" width="204" height="38" rx="19" fill="#06101a" fill-opacity="0.34"/>
  <text x="104" y="643" fill="#f5f8ff" fill-opacity="0.16" font-family="Manrope, system-ui, sans-serif" font-size="20" font-weight="700">${escapeXML(trimForSmallCaption(meta || title))}</text>
</svg>`;
  }

  function buildWideSvg(title, subtitle, palette, seed = 0) {
    const monogram = escapeXML(initials(title).slice(0, 2) || "V");
    return `<svg xmlns="http://www.w3.org/2000/svg" width="640" height="320" viewBox="0 0 640 320">
  <defs>
    <linearGradient id="bg" x1="0" x2="1" y1="0" y2="1">
      <stop offset="0%" stop-color="${palette.base}"/>
      <stop offset="60%" stop-color="${palette.mid}"/>
      <stop offset="100%" stop-color="${palette.edge}"/>
    </linearGradient>
    <radialGradient id="glow" cx="0.85" cy="0.2" r="0.5">
      <stop offset="0%" stop-color="${palette.glow}" stop-opacity="0.5"/>
      <stop offset="100%" stop-color="${palette.glow}" stop-opacity="0"/>
    </radialGradient>
  </defs>
  <rect width="640" height="320" fill="url(#bg)"/>
  <rect width="640" height="320" fill="url(#glow)"/>
  ${wideSceneMarkup(palette, seed)}
  <text x="466" y="184" fill="#f5f8ff" fill-opacity="0.09" font-family="Outfit, Manrope, system-ui, sans-serif" font-size="126" font-weight="800">${monogram}</text>
  <path d="M0 254c130-40 272-54 430-22 72 14 142 38 210 76v12H0z" fill="#071019" fill-opacity="0.42"/>
</svg>`;
  }

  function buildActivitySvg(title, subtitle, palette, seed = 0) {
    const monogram = escapeXML(initials(title).slice(0, 2) || "V");
    return `<svg xmlns="http://www.w3.org/2000/svg" width="132" height="132" viewBox="0 0 132 132">
  <defs>
    <linearGradient id="bg" x1="0" x2="1" y1="0" y2="1">
      <stop offset="0%" stop-color="${palette.base}"/>
      <stop offset="100%" stop-color="${palette.mid}"/>
    </linearGradient>
    <radialGradient id="glow" cx="0.75" cy="0.2" r="0.62">
      <stop offset="0%" stop-color="${palette.glow}" stop-opacity="0.46"/>
      <stop offset="100%" stop-color="${palette.glow}" stop-opacity="0"/>
    </radialGradient>
  </defs>
  <rect width="132" height="132" rx="18" fill="url(#bg)"/>
  <rect width="132" height="132" rx="18" fill="url(#glow)"/>
  ${activitySceneMarkup(palette, seed)}
  <text x="78" y="112" fill="#f5f8ff" fill-opacity="0.16" font-family="Outfit, Manrope, system-ui, sans-serif" font-size="44" font-weight="800">${monogram}</text>
</svg>`;
  }

  function posterSceneMarkup(palette, seed = 0) {
    switch (seed % 4) {
      case 0:
        return `
  <circle cx="462" cy="154" r="118" fill="${palette.glow}" opacity="0.13"/>
  <path d="M392 54 540 184 334 418 188 292z" fill="${palette.accent}" fill-opacity="0.12"/>
  <rect x="84" y="102" width="428" height="476" rx="28" fill="#081017" fill-opacity="0.12"/>
  <rect x="104" y="132" width="132" height="390" rx="20" fill="#ffffff" fill-opacity="0.032"/>
  <rect x="278" y="118" width="184" height="430" rx="24" fill="#ffffff" fill-opacity="0.024"/>`;
      case 1:
        return `
  <ellipse cx="406" cy="212" rx="158" ry="176" fill="${palette.glow}" opacity="0.13"/>
  <rect x="90" y="118" width="164" height="408" rx="24" fill="#ffffff" fill-opacity="0.032"/>
  <rect x="286" y="88" width="188" height="486" rx="28" fill="#ffffff" fill-opacity="0.024"/>
  <path d="M140 548c82-34 154-88 240-144 64-40 126-72 188-86v140H140z" fill="${palette.accent}" fill-opacity="0.11"/>
  <path d="M146 84h320l-70 104H104z" fill="#ffffff" fill-opacity="0.025"/>`;
      case 2:
        return `
  <rect x="70" y="92" width="460" height="524" rx="32" fill="#09121a" fill-opacity="0.12"/>
  <path d="M112 150 254 150 192 502 60 502z" fill="${palette.accent}" fill-opacity="0.12"/>
  <circle cx="446" cy="234" r="86" fill="${palette.glow}" opacity="0.14"/>
  <rect x="278" y="124" width="178" height="386" rx="22" fill="#ffffff" fill-opacity="0.026"/>
  <rect x="104" y="120" width="86" height="306" rx="16" fill="#ffffff" fill-opacity="0.03"/>`;
      default:
        return `
  <circle cx="442" cy="182" r="150" fill="${palette.glow}" opacity="0.13"/>
  <circle cx="442" cy="182" r="62" fill="${palette.accent}" opacity="0.11"/>
  <path d="M170 96h284L322 286H86z" fill="${palette.accent}" fill-opacity="0.12"/>
  <rect x="90" y="112" width="154" height="432" rx="22" fill="#ffffff" fill-opacity="0.032"/>
  <rect x="270" y="140" width="202" height="344" rx="26" fill="#ffffff" fill-opacity="0.024"/>`;
    }
  }

  function heroSceneMarkup(palette, seed = 0) {
    switch (seed % 3) {
      case 0:
        return `
  <path d="M732 74 1126 74 920 412 584 412z" fill="#ffffff" fill-opacity="0.08"/>
  <ellipse cx="930" cy="398" rx="214" ry="256" fill="#09111d" fill-opacity="0.3"/>
  <ellipse cx="1038" cy="378" rx="98" ry="168" fill="#ffffff" fill-opacity="0.05"/>
  <rect x="88" y="122" width="328" height="164" rx="34" fill="#ffffff" fill-opacity="0.03"/>
  <rect x="120" y="162" width="126" height="246" rx="24" fill="#ffffff" fill-opacity="0.03"/>
  <path d="M274 194 468 194 346 414 184 414z" fill="${palette.accent}" fill-opacity="0.13"/>`;
      case 1:
        return `
  <circle cx="980" cy="226" r="178" fill="${palette.glow}" opacity="0.14"/>
  <path d="M678 96c74-36 146-48 252-40 88 6 174 38 262 116v284c-92-88-188-136-282-150-94-14-184 6-274 62z" fill="#0a1524" fill-opacity="0.24"/>
  <rect x="88" y="134" width="288" height="210" rx="30" fill="#ffffff" fill-opacity="0.032"/>
  <rect x="118" y="168" width="116" height="222" rx="22" fill="#ffffff" fill-opacity="0.03"/>
  <path d="M322 112h312l-116 300H198z" fill="${palette.accent}" fill-opacity="0.1"/>`;
      default:
        return `
  <ellipse cx="914" cy="350" rx="238" ry="284" fill="#08111b" fill-opacity="0.26"/>
  <circle cx="1052" cy="198" r="164" fill="${palette.glow}" opacity="0.15"/>
  <path d="M654 92h420L888 470H540z" fill="#ffffff" fill-opacity="0.06"/>
  <rect x="92" y="116" width="354" height="188" rx="34" fill="#ffffff" fill-opacity="0.03"/>
  <rect x="128" y="168" width="148" height="238" rx="24" fill="#ffffff" fill-opacity="0.028"/>
  <path d="M248 182c84-16 172-10 266 16l-114 212H138z" fill="${palette.accent}" fill-opacity="0.11"/>`;
    }
  }

  function wideSceneMarkup(palette, seed = 0) {
    switch (seed % 3) {
      case 0:
        return `
  <path d="M410 10 604 10 482 164 314 164z" fill="${palette.accent}" fill-opacity="0.22"/>
  <rect x="26" y="28" width="208" height="200" rx="22" fill="#ffffff" fill-opacity="0.04"/>
  <rect x="252" y="62" width="238" height="168" rx="22" fill="#ffffff" fill-opacity="0.03"/>
  <rect x="44" y="46" width="78" height="118" rx="16" fill="#ffffff" fill-opacity="0.04"/>`;
      case 1:
        return `
  <circle cx="484" cy="82" r="84" fill="${palette.glow}" opacity="0.18"/>
  <path d="M148 34h284l-120 152H34z" fill="${palette.accent}" fill-opacity="0.18"/>
  <rect x="34" y="46" width="168" height="170" rx="22" fill="#ffffff" fill-opacity="0.035"/>
  <rect x="228" y="54" width="226" height="134" rx="20" fill="#ffffff" fill-opacity="0.03"/>`;
      default:
        return `
  <rect x="26" y="36" width="484" height="178" rx="24" fill="#09111d" fill-opacity="0.18"/>
  <rect x="34" y="44" width="62" height="120" rx="14" fill="#ffffff" fill-opacity="0.04"/>
  <path d="M292 28 462 28 398 154 234 154z" fill="${palette.accent}" fill-opacity="0.22"/>
  <ellipse cx="514" cy="90" rx="72" ry="58" fill="${palette.glow}" opacity="0.18"/>`;
    }
  }

  function activitySceneMarkup(palette, seed = 0) {
    switch (seed % 3) {
      case 0:
        return `<rect x="14" y="16" width="58" height="82" rx="14" fill="#ffffff" fill-opacity="0.05"/><circle cx="102" cy="38" r="18" fill="${palette.accent}" fill-opacity="0.22"/>`;
      case 1:
        return `<path d="M26 18h70L62 88H12z" fill="${palette.accent}" fill-opacity="0.18"/><rect x="18" y="54" width="84" height="48" rx="14" fill="#ffffff" fill-opacity="0.04"/>`;
      default:
        return `<circle cx="90" cy="40" r="24" fill="${palette.glow}" fill-opacity="0.24"/><rect x="18" y="18" width="42" height="76" rx="12" fill="#ffffff" fill-opacity="0.045"/><rect x="64" y="54" width="42" height="32" rx="10" fill="#ffffff" fill-opacity="0.03"/>`;
    }
  }

  function trimForSmallCaption(value = "") {
    const text = normalizeWhitespace(value);
    return text.length > 16 ? `${text.slice(0, 15)}...` : text;
  }

  function wrapTitle(title = "", maxChars = 16, maxLines = 3) {
    const words = normalizeWhitespace(title).split(" ").filter(Boolean);
    if (!words.length) return ["Untitled"];
    const lines = [];
    let current = "";
    while (words.length) {
      const next = words.shift();
      const candidate = current ? `${current} ${next}` : next;
      if (candidate.length <= maxChars || !current) {
        current = candidate;
      } else {
        lines.push(current);
        current = next;
      }
      if (lines.length === maxLines - 1 && words.length) {
        current = `${current} ${words.join(" ")}`.trim();
        words.length = 0;
      }
    }
    if (current) lines.push(current);
    return lines.slice(0, maxLines).map((line, index, output) => index === output.length - 1 && line.length > maxChars + 8 ? `${line.slice(0, maxChars + 5).trimEnd()}...` : line);
  }

  function placeholderPalette(seed = "") {
    const numericSeed = Number.isFinite(Number(seed)) ? Number(seed) : hashString(seed);
    const index = Math.abs(numericSeed) % PLACEHOLDER_PALETTES.length;
    return PLACEHOLDER_PALETTES[index];
  }

  function visualSeed(value = "") {
    return Math.abs(hashString(value || "vyrden-home"));
  }

  function hashString(value = "") {
    let hash = 0;
    for (const char of String(value)) {
      hash = ((hash << 5) - hash) + char.charCodeAt(0);
      hash |= 0;
    }
    return hash;
  }

  function svgDataUrl(svg = "") {
    return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg.replace(/\s{2,}/g, " ").trim())}`;
  }

  function initials(value = "") {
    const words = normalizeWhitespace(value).split(/\s+/).filter(Boolean);
    if (!words.length) return "V";
    if (words.length === 1) return words[0].slice(0, 1).toUpperCase();
    return `${words[0][0] || ""}${words[1][0] || ""}`.toUpperCase();
  }

  function firstNonEmpty() {
    for (const value of arguments) {
      const text = normalizeWhitespace(value);
      if (text) return text;
    }
    return "";
  }

  function clampPercent(value) {
    const percent = Number(value || 0);
    if (!Number.isFinite(percent)) return 0;
    if (percent <= 1) return Math.max(0, Math.min(100, Math.round(percent * 100)));
    return Math.max(0, Math.min(100, Math.round(percent)));
  }

  function escapeRegex(value = "") {
    return String(value).replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  }

  return {
    buildPlaceholderArtwork,
    cleanDescription,
    cleanDisplayMeta,
    cleanDisplaySubtitle,
    cleanDisplayTitle,
    escapeHTML,
    extractEpisodeCode,
    extractYear,
    kindLabel,
    presentHomeItem,
  };
});
