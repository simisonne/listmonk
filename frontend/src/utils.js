import {
  DialogProgrammatic as Dialog,
  ToastProgrammatic as Toast,
} from 'buefy';
import dayjs from 'dayjs';
import dayDuration from 'dayjs/plugin/duration';
import relativeTime from 'dayjs/plugin/relativeTime';
import updateLocale from 'dayjs/plugin/updateLocale';

dayjs.extend(updateLocale);
dayjs.extend(relativeTime);
dayjs.extend(dayDuration);

const reEmail = /(.+?)@(.+?)/ig;
const prefKey = 'listmonk_pref';

// Campaign palette. The order is load bearing: the SQL fallback in
// internal/migrations/v6.7.0.go picks the entry at (id % N) from this exact
// list, so both sides have to stay identical entry for entry.
export const CAMPAIGN_PALETTE = [
  {
    key: 'tt-onboard', base: '#14B8A6', tint: '#CCFBF1', deep: '#0F766E',
  },
  {
    key: 'tt-one-off', base: '#6366F1', tint: '#E0E7FF', deep: '#4338CA',
  },
  {
    key: 'ig-onboard', base: '#EC4899', tint: '#FCE7F3', deep: '#BE185D',
  },
  {
    key: 'email-list', base: '#3E6FBF', tint: '#DBEAFE', deep: '#1E4FA8',
  },
  {
    key: 'bs-loopkit', base: '#F59E0B', tint: '#FEF3C7', deep: '#B45309',
  },
  {
    key: 'emerald', base: '#10B981', tint: '#D1FAE5', deep: '#047857',
  },
  {
    key: 'violet', base: '#8B5CF6', tint: '#EDE9FE', deep: '#6D28D9',
  },
  {
    key: 'cyan', base: '#0891B2', tint: '#CFFAFE', deep: '#0E7490',
  },
];

export const EVENT_COLORS = {
  melodies: { base: '#65A30D', deep: '#3F6212' },
  negative: { base: '#EF4444', deep: '#B91C1C' },
};

// #rrggbb (or #rgb) hex plus an alpha as an rgba() string, for inline styles.
const withAlpha = (hex, alpha) => {
  let h = String(hex || '').replace('#', '');
  if (h.length === 3) {
    h = h.split('').map((c) => c + c).join('');
  }
  if (h.length !== 6 || /[^0-9a-f]/i.test(h)) {
    return `rgba(0, 0, 0, ${alpha})`;
  }
  const r = parseInt(h.slice(0, 2), 16);
  const g = parseInt(h.slice(2, 4), 16);
  const b = parseInt(h.slice(4, 6), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
};

const htmlEntities = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;',
  "'": '&#39;',
  '/': '&#x2F;',
  '`': '&#x60;',
  '=': '&#x3D;',
};

export default class Utils {
  constructor(i18n) {
    this.i18n = i18n;
    this.intlNumFormat = new Intl.NumberFormat();

    if (i18n) {
      dayjs.updateLocale('en', {
        relativeTime: {
          future: '%s',
          past: '%s',
          s: `${i18n.tc('globals.terms.second', 2)}`,
          m: `1 ${i18n.tc('globals.terms.minute', 1)}`,
          mm: `%d ${i18n.tc('globals.terms.minute', 2)}`,
          h: `1 ${i18n.tc('globals.terms.hour', 1)}`,
          hh: `%d ${i18n.tc('globals.terms.hour', 2)}`,
          d: `1 ${i18n.tc('globals.terms.day', 1)}`,
          dd: `%d ${i18n.tc('globals.terms.day', 2)}`,
          M: `1 ${i18n.tc('globals.terms.month', 1)}`,
          MM: `%d ${i18n.tc('globals.terms.month', 2)}`,
          y: `${i18n.tc('globals.terms.year', 1)}`,
          yy: `%d ${i18n.tc('globals.terms.year', 2)}`,
        },
      });
    }
  }

  getDate = (d) => dayjs(d);

  // Parses an ISO timestamp to a simpler form.
  niceDate = (stamp, showTime) => {
    if (!stamp) {
      return '';
    }

    const d = dayjs(stamp);
    const day = this.i18n.t(`globals.days.${d.day() + 1}`);
    const month = this.i18n.t(`globals.months.${d.month() + 1}`);
    let out = d.format(`[${day},] DD [${month}] YYYY`);
    if (showTime) {
      out += d.format(', HH:mm');
    }

    return out;
  };

  duration = (start, end) => {
    const a = dayjs(start);
    const b = dayjs(end);
    const d = dayjs.duration(Math.abs(b.diff(a)));

    const parts = [
      Math.floor(d.asDays()) && `${Math.floor(d.asDays())}d`,
      d.hours() && `${d.hours()}h`,
      d.minutes() && `${d.minutes()}m`,
      d.seconds() && `${d.seconds()}s`,
    ].filter(Boolean);

    return `${b.isBefore(a) ? '-' : ''}${parts.join(' ')}`;
  };

  // Simple, naive, e-mail address check.
  validateEmail = (e) => e.match(reEmail);

  niceNumber = (n) => {
    if (n === null || n === undefined) {
      return 0;
    }

    let pfx = '';
    let div = 1;

    if (n >= 1.0e+9) {
      pfx = 'b';
      div = 1.0e+9;
    } else if (n >= 1.0e+6) {
      pfx = 'm';
      div = 1.0e+6;
    } else if (n >= 1.0e+4) {
      pfx = 'k';
      div = 1.0e+3;
    } else {
      return n;
    }

    // Whole number without decimals.
    const out = (n / div);
    if (Math.floor(out) === n) {
      return out + pfx;
    }

    return out.toFixed(2) + pfx;
  };

  formatNumber(v) {
    return this.intlNumFormat.format(v);
  }

  // Restore HTML-escaped quotes (&quot;, &#34;, &#x22;) inside {{ ... }}
  // template actions back to plain quotes. The richtext editor serializes
  // attribute values with double quotes, so href='{{ TrackLink "..." }}'
  // comes back as href="{{ TrackLink &quot;...&quot; }}", and Go templates
  // fail that with: unexpected "&" in operand. Text outside actions is
  // left untouched.
  fixMangledTemplateQuotes(body) {
    if (!body) {
      return body;
    }
    return body.replace(/{{(.*?)}}/gs, (m, inner) => `{{${inner.replace(/&quot;|&#0*34;|&#[xX]0*22;/gi, '"')}}}`);
  }

  // Parse one or more numeric ids as query params and return as an array of ints.
  parseQueryIDs = (ids) => {
    if (!ids) {
      return [];
    }

    if (typeof ids === 'string') {
      return [parseInt(ids, 10)];
    }

    if (typeof ids === 'number') {
      return [parseInt(ids, 10)];
    }

    return ids.map((id) => parseInt(id, 10));
  };

  // https://stackoverflow.com/a/12034334
  escapeHTML = (html) => html.replace(/[&<>"'`=/]/g, (s) => htmlEntities[s]);

  titleCase = (str) => str[0].toUpperCase() + str.substr(1).toLowerCase();

  // UI shortcuts.
  confirm = (msg, onConfirm, onCancel) => {
    Dialog.confirm({
      scroll: 'keep',
      message: !msg ? this.i18n.t('globals.messages.confirm') : this.escapeHTML(msg),
      confirmText: this.i18n.t('globals.buttons.ok'),
      cancelText: this.i18n.t('globals.buttons.cancel'),
      onConfirm,
      onCancel,
    });
  };

  prompt = (msg, inputAttrs, onConfirm, onCancel, params) => {
    const p = params || {};

    Dialog.prompt({
      scroll: 'keep',
      message: this.escapeHTML(msg),
      confirmText: p.confirmText || this.i18n.t('globals.buttons.ok'),
      cancelText: p.cancelText || this.i18n.t('globals.buttons.cancel'),
      inputAttrs: {
        type: 'string',
        maxlength: 200,
        ...inputAttrs,
      },
      trapFocus: true,
      onConfirm,
      onCancel,
    });
  };

  toast = (msg, typ, duration, queue) => {
    Toast.open({
      message: this.escapeHTML(msg),
      type: !typ ? 'is-success' : typ,
      queue,
      duration: duration || 3000,
      position: 'is-top',
      pauseOnHover: true,
    });
  };

  // Takes a props.row from a Buefy b-column <td> template and
  // returns a `data-id` attribute which Buefy then applies to the td.
  tdID = (row) => ({ 'data-id': row.id.toString() });

  camelString = (str) => {
    const s = str.replace(/[-_\s]+(.)?/g, (match, chr) => (chr ? chr.toUpperCase() : ''));
    return s.substr(0, 1).toLowerCase() + s.substr(1);
  };

  // camelKeys recursively camelCases all keys in a given object (array or {}).
  // For each key it traverses, it passes a dot separated key path to an optional testFunc() bool.
  // so that it can camelcase or leave a particular key alone based on what testFunc() returns.
  // eg: The keypath for {"data": {"results": ["created_at": 123]}} is
  // .data.results.*.created_at (array indices become *)
  // testFunc() can examine this key and return true to convert it to camelcase
  // or false to leave it as-is.
  camelKeys = (obj, testFunc, keys) => {
    if (obj === null) {
      return obj;
    }

    if (Array.isArray(obj)) {
      return obj.map((o) => this.camelKeys(o, testFunc, `${keys || ''}.*`));
    }

    if (obj.constructor === Object) {
      return Object.keys(obj).reduce((result, key) => {
        const keyPath = `${keys || ''}.${key}`;
        let k = key;

        // If there's no testfunc or if a function is defined and it returns true, convert.
        if (testFunc === undefined || testFunc(keyPath)) {
          k = this.camelString(key);
        }

        return {
          ...result,
          [k]: this.camelKeys(obj[key], testFunc, keyPath),
        };
      }, {});
    }

    return obj;
  };

  // Palette entry for a campaign. An explicit campaigns.color that is in the
  // palette wins, any other stored colour is used as-is (washed out for the
  // highlight background), and an empty colour falls back to the id % N entry that
  // mirrors the SQL backfill in internal/migrations/v6.7.0.go.
  campaignPalette = (campaign) => {
    const color = String((campaign && campaign.color) || '').trim().toLowerCase();
    const id = parseInt((campaign && campaign.id), 10);

    if (color) {
      const hit = CAMPAIGN_PALETTE.find((p) => p.base === color);
      if (hit) {
        return { ...hit, known: true };
      }
      return {
        key: 'custom', base: color, tint: withAlpha(color, 0.15), deep: color, known: false,
      };
    }

    const n = CAMPAIGN_PALETTE.length;
    const idx = Number.isNaN(id) ? 0 : ((id % n) + n) % n;
    return { ...CAMPAIGN_PALETTE[idx], known: true };
  };

  // Inline style for the highlight behind every campaign name: the saturated
  // base colour at low alpha so the page background shows through, deep text
  // colour, no border.
  campaignChipStyle = (campaign) => {
    const p = this.campaignPalette(campaign);
    return {
      backgroundColor: withAlpha(p.base, 0.10),
      color: p.deep,
    };
  };

  // Highlight behind non-campaign event text in the dashboard feed and behind
  // 'melodies' covers Melodies site activity and list optins
  // (lime, used by no campaign type and far from every palette hue), 'negative' covers
  // unsubscribes and bounces (red). Same low alpha as campaign names.
  eventHighlightStyle = (key) => {
    const c = EVENT_COLORS[key] || EVENT_COLORS.melodies;
    return {
      backgroundColor: withAlpha(c.base, 0.10),
      color: c.deep,
    };
  };

  // Inline style for one swatch in the campaign colour picker: the same
  // recipe as the highlight (tint background, deep text, deep border at 25% alpha).
  // No palette entry means "auto", which stays neutral grey.
  campaignSwatchStyle = (paletteEntry) => {
    if (!paletteEntry) {
      return {
        backgroundColor: '#f5f5f5',
        color: '#4a4a4a',
        borderColor: withAlpha('#4a4a4a', 0.25),
      };
    }
    return {
      backgroundColor: paletteEntry.tint,
      color: paletteEntry.deep,
      borderColor: withAlpha(paletteEntry.deep, 0.25),
    };
  };

  getPref = (key) => {
    if (localStorage.getItem(prefKey) === null) {
      return null;
    }

    const p = JSON.parse(localStorage.getItem(prefKey));
    return key in p ? p[key] : null;
  };

  setPref = (key, val) => {
    let p = {};
    if (localStorage.getItem(prefKey) !== null) {
      p = JSON.parse(localStorage.getItem(prefKey));
    }

    p[key] = val;
    localStorage.setItem(prefKey, JSON.stringify(p));
  };
}
