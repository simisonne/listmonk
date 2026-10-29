<template>
  <section class="website content">
    <header class="columns">
      <div class="column is-two-thirds">
        <h1 class="title is-5">Website</h1>
      </div>
    </header>

    <article class="notification">
      <div class="filters">
        <div class="select is-small">
          <label class="is-sr-only" for="site-filter">Site</label>
          <select id="site-filter" v-model="siteFilter" data-cy="site-filter">
            <option value="">All sites ({{ shownByHidden.length }})</option>
            <option v-for="o in siteOptions" :key="o.key" :value="o.key">
              {{ o.label }} ({{ o.count }})
            </option>
          </select>
        </div>

        <b-switch v-model="showHidden" size="is-small" data-cy="show-hidden">
          Show visits from Regensburg
        </b-switch>

        <span class="has-text-grey is-size-7">{{ visible.length }} of {{ shownByHidden.length }} events</span>
      </div>

      <div class="relative">
        <b-loading v-if="isLoading" active :is-full-page="false" />
        <div class="table-container" v-if="visible.length">
          <table class="table is-fullwidth is-hoverable site-table">
            <thead>
              <tr>
                <th>When</th>
                <th>Site</th>
                <th>Event</th>
                <th>Detail</th>
                <th>Device</th>
                <th>Location</th>
                <th>IP</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(e, i) in visible" :key="eventKey(e, i)" :class="{ 'is-own': e.hidden }">
                <td class="when" :title="relative(e.createdAt)">{{ when(e.createdAt) }}</td>
                <td>
                  <span class="site-chip" :style="chipStyle(e)">{{ siteLabel(e) }}</span>
                </td>
                <td>{{ eventLabel(e) }}</td>
                <td class="detail" :title="detail(e)">{{ detail(e) }}</td>
                <td>{{ device(e) }}</td>
                <td>{{ location(e) }}</td>
                <td class="ip">{{ e.ip || '' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else-if="!isLoading" class="has-text-grey">
          No events for this filter.
        </p>
      </div>
    </article>
  </section>
</template>

<script>
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import Vue from 'vue';
import {
  SITE_LABELS, eventSiteKey, siteColorKey,
} from '../utils';

dayjs.extend(relativeTime);

// The site event types this page lists, in the words a visitor's action gets.
const EVENT_LABELS = {
  site_visit: 'Page visit',
  beats_visit: 'Page visit',
  weekly_loops_page: 'Page visit',
  portfolio_visit: 'Page visit',
  portfolio_referral: 'Referral visit',
  track_played: 'Play',
  track_downloaded: 'Download',
};

export default Vue.extend({
  name: 'Website',

  data() {
    return {
      isLoading: true,
      events: [],
      siteFilter: '',
      // Regensburg is Simon's own egress: it stays out of the list until the
      // switch turns it on. The server sends those rows flagged, so toggling
      // costs nothing.
      showHidden: false,
    };
  },

  methods: {
    fetchData() {
      this.isLoading = true;
      this.$api.getSiteEvents(500).then((data) => {
        this.events = (data || []).slice().sort((a, b) => String(b.createdAt).localeCompare(String(a.createdAt)));
        this.isLoading = false;
      });
    },

    eventKey(e, i) {
      return `${i}-${e.type}-${e.createdAt}-${e.ip || ''}`;
    },

    // Which site a row came from, in the vocabulary the filters use.
    siteKey(e) {
      return eventSiteKey(e);
    },

    siteLabel(e) {
      return SITE_LABELS[this.siteKey(e)] || SITE_LABELS.melodies;
    },

    chipStyle(e) {
      return this.$utils.eventHighlightStyle(siteColorKey(this.siteKey(e)));
    },

    eventLabel(e) {
      return EVENT_LABELS[e.type] || e.type;
    },

    // What the row was about: the track, the page path, or the referral tag.
    detail(e) {
      if (e.track) {
        return e.track;
      }
      if (e.ref) {
        return `@${String(e.ref).replace(/^@/, '')}`;
      }
      if (e.path) {
        return e.path;
      }
      return '';
    },

    // Exactly what the log recorded, the same way the dashboard feed reads
    // it: "Bot, Bot" and "Unknown, Unknown" are information here, an empty
    // cell would look like the row had no device at all.
    device(e) {
      return [e.device, e.browser].filter((v) => v).join(', ');
    },

    location(e) {
      return this.$utils.prettyLocation(e.location);
    },

    when(stamp) {
      return stamp ? dayjs(stamp).format('DD MMM HH:mm') : '';
    },

    relative(stamp) {
      return stamp ? dayjs(stamp).fromNow() : '';
    },

    onRefresh() {
      this.fetchData();
    },
  },

  computed: {
    // Everything the Regensburg switch lets through.
    shownByHidden() {
      if (this.showHidden) {
        return this.events;
      }
      return this.events.filter((e) => !e.hidden);
    },

    visible() {
      if (!this.siteFilter) {
        return this.shownByHidden;
      }
      return this.shownByHidden.filter((e) => this.siteKey(e) === this.siteFilter);
    },

    // Every tracked site stays in the picker, with the count of what the
    // switch currently lets through, so the list does not reshuffle.
    siteOptions() {
      return Object.keys(SITE_LABELS).map((key) => ({
        key,
        label: SITE_LABELS[key],
        count: this.shownByHidden.filter((e) => this.siteKey(e) === key).length,
      }));
    },
  },

  created() {
    this.$root.$on('page.refresh', this.onRefresh);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.onRefresh);
  },

  mounted() {
    this.fetchData();
  },
});
</script>

<style scoped>
.filters {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.75rem;
}
.relative {
  position: relative;
  min-height: 6rem;
}
.site-table {
  font-size: 0.85rem;
}
.site-table td {
  padding: 0.3rem 0.5rem;
  vertical-align: middle;
}
.site-table th {
  padding: 0.3rem 0.5rem;
  font-size: 0.8rem;
  color: #7a7a7a;
  font-weight: 600;
}
.site-chip {
  display: inline-block;
  padding: 0 0.45rem;
  border-radius: 3px;
  font-weight: 700;
  white-space: nowrap;
}
.site-table .when {
  white-space: nowrap;
  color: #4a4a4a;
}
.site-table .detail {
  max-width: 22rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.site-table .ip {
  font-family: monospace;
  font-size: 0.78rem;
  color: #4a4a4a;
}
/* Simon's own network: kept in the list only when the switch is on. */
.site-table tr.is-own {
  opacity: 0.55;
}
</style>
