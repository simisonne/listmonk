<template>
  <section class="dashboard content">
    <header class="columns">
      <div class="column is-two-thirds">
        <h1 class="title is-5">
          {{ $utils.niceDate(new Date()) }}
        </h1>
      </div>
    </header>

    <!-- Loud, unmissable alert for campaigns that ended without reaching
         every recipient. listmonk marks such a campaign finished even when
         most messages failed, so without this the loss is invisible. -->
    <article v-if="partialCampaigns.length" class="partial-alert" data-cy="partial-campaigns">
      <div class="partial-alert-head">
        <b-icon icon="alert-octagon" />
        <span>
          {{ partialCampaigns.length }}
          campaign{{ partialCampaigns.length > 1 ? 's' : '' }}
          did not reach every recipient
        </span>
      </div>
      <ul class="partial-alert-list">
        <li v-for="c in partialCampaigns" :key="c.id">
          <router-link :to="`/campaigns/${c.id}`" class="partial-name">
            {{ c.name }}
          </router-link>
          <span class="partial-count">
            {{ $utils.niceNumber(c.sent) }} of {{ $utils.niceNumber(c.toSend) }} sent,
            <strong>{{ $utils.niceNumber(c.missing) }} missing</strong>
          </span>
          <span class="partial-when">{{ partialWhen(c) }}</span>
          <a href="#" class="partial-dismiss" @click.prevent="dismissPartial(c)">
            handled, hide
          </a>
        </li>
      </ul>
      <p class="partial-help">
        These sends stopped early and no address bounced. Find out why in the log,
        send the missing recipients again, then hide the row.
      </p>
    </article>

    <section class="counts wrap">
      <div class="tile is-ancestor">
        <div class="tile is-vertical is-12">
          <div class="tile">
            <div class="tile is-parent is-vertical relative">
              <b-loading v-if="isCountsLoading" active :is-full-page="false" />
              <article class="tile is-child notification" data-cy="lists">
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="format-list-bulleted-square" />
                      {{ $utils.niceNumber(counts.lists.total) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $tc('globals.terms.list', counts.lists.total) }}
                    </p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.public) }}</label>
                        {{ $t('lists.types.public') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.private) }}</label>
                        {{ $t('lists.types.private') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.optinSingle) }}</label>
                        {{ $t('lists.optins.single') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.optinDouble) }}</label>
                        {{ $t('lists.optins.double') }}
                      </li>
                    </ul>
                  </div>
                </div>
              </article><!-- lists -->

              <article class="tile is-child notification" data-cy="campaigns">
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="rocket-launch-outline" />
                      {{ $utils.niceNumber(counts.campaigns.total) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $tc('globals.terms.campaign', counts.campaigns.total) }}
                    </p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li v-for="(num, status) in counts.campaigns.byStatus" :key="status">
                        <label for="#" :data-cy="`campaigns-${status}`">{{ num }}</label>
                        {{ $t(`campaigns.status.${status}`) }}
                        <span v-if="status === 'running'" class="spinner is-tiny">
                          <b-loading :is-full-page="false" active />
                        </span>
                      </li>
                    </ul>
                  </div>
                </div>
              </article><!-- campaigns -->
            </div><!-- block -->

            <div class="tile is-parent relative">
              <b-loading v-if="isCountsLoading" active :is-full-page="false" />
              <article class="tile is-child notification" data-cy="subscribers">
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="account-multiple" />
                      {{ $utils.niceNumber(counts.subscribers.total) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $tc('globals.terms.subscriber', counts.subscribers.total) }}
                    </p>
                  </div>

                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.subscribers.blocklisted) }}</label>
                        {{ $t('subscribers.status.blocklisted') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.subscribers.orphans) }}</label>
                        {{ $t('dashboard.orphanSubs') }}
                      </li>
                    </ul>
                  </div><!-- subscriber breakdown -->
                </div><!-- subscriber columns -->
                <hr />
                <div class="columns is-mobile" data-cy="messages">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="email-outline" />
                      {{ $utils.niceNumber(counts.messages.total) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $t('dashboard.messagesSent') }}
                    </p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.messages.opens) }}</label>
                        {{ $t('dashboard.opens') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.messages.clicks) }}</label>
                        {{ $t('dashboard.clicks') }}
                      </li>
                    </ul>
                  </div>
                </div>
              </article><!-- subscribers -->
            </div>
          </div>
          <div class="tile is-parent relative">
            <b-loading v-if="isChartsLoading" active :is-full-page="false" />
            <article class="tile is-child notification charts">
              <div class="columns">
                <div class="column is-6">
                  <h3 class="title is-size-6">
                    {{ $t('dashboard.campaignViews') }}
                  </h3><br />
                  <chart type="line" v-if="campaignViews" :data="campaignViews" />
                </div>
                <div class="column is-6">
                  <h3 class="title is-size-6 has-text-right">
                    {{ $t('dashboard.linkClicks') }}
                  </h3><br />
                  <chart type="line" v-if="campaignClicks" :data="campaignClicks" />
                </div>
              </div>
            </article>
          </div>
          <div class="tile is-parent relative">
            <b-loading v-if="isEventsLoading" active :is-full-page="false" />
            <article class="tile is-child notification" data-cy="events">
              <h3 class="title is-size-6">
                {{ $t('dashboard.recentEvents') }}
              </h3>
              <ul v-if="visibleEvents.length" class="events">
                <li v-for="(e, i) in visibleEvents" :key="eventKey(e, i)" class="event" data-cy="event">
                  <small class="has-text-grey timestamp" :title="eventRelative(e.createdAt)">{{ eventTime(e.createdAt) }}</small>
                  <b-icon :icon="eventIcon(e.type)" size="is-small" />
                  <span class="event-text">
                    <template v-if="e.type !== 'campaign_sent'">
                      <template v-if="e.email && e.subscriberId && !isMelodiesEvent(e.type)">
                        <router-link :to="{ name: 'subscriber', params: { id: e.subscriberId }, query: { tab: 'activity' } }">{{ e.email }}</router-link>
                      </template>
                      <template v-else-if="e.email">{{ e.email }}</template>
                      <template v-else-if="e.subscriberName">{{ e.subscriberName }}</template>
                      <template v-else>Someone</template>
                      {{ ' ' }}
                    </template>
                    <template v-if="e.type === 'campaign_sent'">
                      <template v-if="e.campaignId">
                        Campaign sent:
                        <router-link :to="{ name: 'campaignAnalytics', query: { id: e.campaignId } }">
                          <span class="campaign-name-hl" :style="campChip(e)">{{ e.campaignName || `#${e.campaignId}` }}</span>
                        </router-link>
                      </template>
                      <template v-else>
                        Campaign sent: <span class="campaign-name-hl" :style="campChip(e)">{{ e.campaignName || 'a campaign' }}</span>
                      </template>
                    </template>
                    <template v-else-if="e.type === 'open'">
                      opened
                      <router-link v-if="e.campaignId" :to="{ name: 'campaignAnalytics', query: { id: e.campaignId } }">
                        <span class="campaign-name-hl" :style="campChip(e)">{{ e.campaignName || 'a campaign' }}</span>
                      </router-link>
                      <template v-else>
                        <span class="campaign-name-hl" :style="campChip(e)">{{ e.campaignName || 'a campaign' }}</span>
                      </template>
                      <template v-if="e.openCount > 1"> ({{ ordinal(e.openCount) }} time)</template>
                    </template>
                    <template v-else-if="e.type === 'click'">
                      clicked a link in
                      <router-link v-if="e.campaignId" :to="{ name: 'campaignAnalytics', query: { id: e.campaignId } }">
                        <span class="campaign-name-hl" :style="campChip(e)">{{ e.campaignName || 'a campaign' }}</span>
                      </router-link>
                      <template v-else>
                        <span class="campaign-name-hl" :style="campChip(e)">{{ e.campaignName || 'a campaign' }}</span>
                      </template>
                      <template v-if="e.openCount > 1"> ({{ ordinal(e.openCount) }} time)</template>
                    </template>
                    <template v-else-if="e.type === 'optin'">joined <span class="campaign-name-hl" :style="eventHl('melodies')">{{ e.listName || 'a public list' }}</span></template>
                    <template v-else-if="e.type === 'unsubscribe'">left <span class="campaign-name-hl" :style="eventHl('negative')">{{ e.listName || 'a list' }}</span></template>
                    <template v-else-if="e.type === 'site_visit'"><span class="campaign-name-hl" :style="eventHl('melodies')">visited the Melodies site</span>{{ melodiesDetail(e) }}</template>
                    <template v-else-if="e.type === 'beats_visit'"><span class="campaign-name-hl" :style="eventHl('beats')">visited the Beats site</span>{{ melodiesDetail(e) }}</template>
                    <template v-else-if="e.type === 'track_played'"><span class="campaign-name-hl"
                      :style="eventHl(eventSiteKey(e))">played{{ e.track ? `: ${e.track}` : ' a track' }} on {{ siteName(e) }}</span>{{ melodiesDetail(e) }}</template>
                    <template v-else-if="e.type === 'track_downloaded'"><span class="campaign-name-hl"
                      :style="eventHl(eventSiteKey(e))">downloaded{{ e.track ? `: ${e.track}` : ' a track' }} from {{ siteName(e) }}</span>{{ melodiesDetail(e) }}</template>
                    <template v-else-if="e.type === 'portfolio_visit'"><span class="campaign-name-hl"
                      :style="eventHl('portfolio')">visited the portfolio site{{ pageSuffix(e) }}</span>{{ siteDetail(e) }}</template>
                    <template v-else-if="e.type === 'portfolio_referral'"><span class="campaign-name-hl"
                      :style="eventHl('referral')">visited the portfolio site via your @{{ refName(e) }} link</span>{{ siteDetail(e) }}</template>
                    <template v-else-if="e.type === 'weekly_loops_page'"><span class="campaign-name-hl"
                      :style="eventHl('weekly')">visited the weekly loops page</span>{{ melodiesDetail(e) }}</template>
                    <template v-else>{{ e.type }}</template>
                  </span>
                  <button type="button" class="event-del" title="Hide from recent events, nothing is deleted"
                    aria-label="Hide from recent events" @click.prevent="hideEvent(e)">
                    <b-icon icon="trash-can-outline" size="is-small" />
                  </button>
                </li>
              </ul>
              <p v-else-if="!isEventsLoading" class="has-text-grey">
                {{ $t('dashboard.noEvents') }}
              </p>
                <a v-if="activeEvents.length > 10" href="#" class="toggle"
                  @click.prevent="eventsExpanded = !eventsExpanded">
                  {{ eventsExpanded ? $t('dashboard.showLess') : $t('dashboard.showMore') }}
                </a>
            </article>
          </div>
        </div>
      </div><!-- tile block -->
      <p v-if="settings['app.cache_slow_queries']" class="has-text-grey">
        *{{ $t('globals.messages.slowQueriesCached') }}
        <a href="https://listmonk.app/docs/maintenance/performance/" target="_blank" rel="noopener noreferer"
          class="has-text-grey">
          <b-icon icon="link-variant" /> {{ $t('globals.buttons.learnMore') }}
        </a>
      </p>
    </section>
  </section>
</template>

<script>
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import Vue from 'vue';
import { mapState } from 'vuex';
import { colors } from '../constants';
import { eventSiteKey as siteKeyOf, siteColorKey } from '../utils';
import Chart from '../components/Chart.vue';

dayjs.extend(relativeTime);

// Where this browser remembers the feed rows Simon hid. Hiding is a display
// choice only: nothing is deleted from the DB, the activity log keeps its
// lines, and the Website page still lists every visit.
const HIDDEN_EVENTS_KEY = 'dashboard.hiddenEvents';

// Hidden rows whose key has not been seen for this long are forgotten. They
// scrolled out of the fetched window long ago and can never come back.
const HIDDEN_EVENTS_MAX_AGE = 1000 * 60 * 60 * 24 * 90;

export default Vue.extend({
  components: {
    Chart,
  },

  data() {
    return {
      isChartsLoading: true,
      isCountsLoading: true,
      isEventsLoading: true,
      eventsExpanded: false,
      events: [],
      hiddenEvents: {},
      partialCampaigns: [],
      campaignViews: null,
      campaignClicks: null,
      refreshTimer: null,
      counts: {
        lists: {},
        subscribers: {},
        campaigns: {},
        messages: {},
      },
    };
  },

  methods: {
    fetchData(quiet) {
      if (!quiet) {
        this.isCountsLoading = true;
        this.isChartsLoading = true;
      }

      this.$api.getDashboardCounts().then((data) => {
        this.counts = data;
        this.isCountsLoading = false;
      });

      this.$api.getDashboardCharts().then((data) => {
        this.isChartsLoading = false;
        this.campaignViews = this.makeChart(data.campaignViews);
        this.campaignClicks = this.makeChart(data.linkClicks);
      });

      // Fetch the 80 newest events once; the card shows 10 until expanded.
      if (!quiet) {
        this.isEventsLoading = true;
      }
      this.$api.getDashboardEvents(80).then((data) => {
        this.events = data;
        this.isEventsLoading = false;
      });

      this.$api.getPartialCampaigns().then((data) => {
        this.partialCampaigns = data || [];
      });
    },

    // partialWhen renders when a partial campaign last moved.
    partialWhen(c) {
      return dayjs(c.updatedAt).fromNow();
    },

    // dismissPartial hides a partial campaign from the alert permanently.
    dismissPartial(c) {
      this.$api.dismissPartialCampaign(c.id).then(() => {
        this.partialCampaigns = this.partialCampaigns.filter((x) => x.id !== c.id);
        this.$utils.toast('Hidden. That campaign will not be alerted again.');
      });
    },

    // Identity of one feed row. Neither the DB half nor the log half of the
    // feed carries an id, so the row's own values stand in for one: two rows
    // that agree on all of them are the same row for hiding purposes.
    eventKey(e) {
      return [
        e.type, e.createdAt, e.campaignId, e.subscriberId, e.email, e.listName,
        e.url, e.track, e.openCount, e.site, e.ip, e.path, e.ref,
      ].join('|');
    },

    loadHiddenEvents() {
      const now = Date.now();
      const kept = {};
      try {
        const raw = window.localStorage.getItem(HIDDEN_EVENTS_KEY);
        const parsed = raw ? JSON.parse(raw) : {};
        Object.keys(parsed).forEach((k) => {
          if (now - parsed[k] < HIDDEN_EVENTS_MAX_AGE) {
            kept[k] = parsed[k];
          }
        });
      } catch (err) {
        // Unreadable or blocked storage just means nothing is hidden yet.
      }
      this.hiddenEvents = kept;
    },

    persistHiddenEvents() {
      try {
        window.localStorage.setItem(HIDDEN_EVENTS_KEY, JSON.stringify(this.hiddenEvents));
      } catch (err) {
        // Out of quota or storage blocked: the hiding still holds for this page.
      }
    },

    // Hiding has no confirm step on purpose: the row only leaves the feed and
    // nothing is deleted. There is no undo in the card, so a hidden row stays
    // hidden in this browser; clear the dashboard.hiddenEvents storage key to
    // bring them all back.
    hideEvent(e) {
      this.hiddenEvents = { ...this.hiddenEvents, [this.eventKey(e)]: Date.now() };
      this.persistHiddenEvents();
      this.$utils.toast('Hidden from recent events. Nothing was deleted.');
    },

    makeChart(data) {
      if (data.length === 0) {
        return {};
      }
      return {
        labels: data.map((d) => dayjs(d.date).format('DD MMM')),
        datasets: [
          {
            data: [...data.map((d) => d.count)],
            borderColor: colors.primary,
            borderWidth: 2,
            pointHoverBorderWidth: 5,
            pointBorderWidth: 0.5,
          },
        ],
      };
    },

    eventIcon(type) {
      return {
        campaign_sent: 'rocket-launch-outline',
        open: 'email-open-outline',
        click: 'cursor-default-click-outline',
        optin: 'account-plus-outline',
        unsubscribe: 'account-minus-outline',
        site_visit: 'web',
        beats_visit: 'web',
        track_played: 'music',
        track_downloaded: 'download',
        portfolio_visit: 'web',
        portfolio_referral: 'link-variant',
        weekly_loops_page: 'calendar-clock',
      }[type] || 'bell-outline';
    },

    // Inline style for the highlight behind a campaign name in an event.
    campChip(e) {
      return this.$utils.campaignChipStyle({ id: e.campaignId, color: e.campaignColor });
    },

    // Inline style for the highlight behind a non-campaign event description.
    eventHl(key) {
      return this.$utils.eventHighlightStyle(key);
    },

    // Melodies site events stay as plain text: nothing in them is clickable.
    isMelodiesEvent(type) {
      return ['site_visit', 'beats_visit', 'track_played', 'track_downloaded',
        'weekly_loops_page'].indexOf(type) !== -1;
    },

    // " (iPhone, Safari) from Berlin, DE" suffix for melodies events,
    // empty when the log line predates the device / browser fields and
    // carries no (or a private) location.
    melodiesDetail(e) {
      const parts = [e.device, e.browser].filter((v) => v);
      let out = parts.length ? ` (${parts.join(', ')})` : '';
      const loc = this.$utils.prettyLocation(e.location);
      if (loc) {
        out += ` from ${loc}`;
      }
      return out;
    },

    // Which of the site pages a track event came from, as a word for the
    // sentence. Log lines written before the site= tag existed, and the
    // older guest pages, carry nothing and keep the Melodies wording they
    // have always been rendered with.
    siteName(e) {
      const key = String(e.site || '').toLowerCase();
      if (key === 'beats') return 'the Beats site';
      if (key === 'loops') return 'the Loops site';
      return 'the Melodies site';
    },

    // The matching highlight colour key for eventHl(): the loops site wears
    // the teal its page visits use.
    eventSiteKey(e) {
      return siteColorKey(siteKeyOf(e));
    },

    // ": education" page suffix for portfolio visits, empty for the home
    // page or when the log line carries no path.
    pageSuffix(e) {
      const p = e.path || '';
      if (!p || p === '/hire-me/' || p === '/hire-me/index.html' || p === 'index.html') {
        return '';
      }
      const clean = p.replace(/^\/hire-me\//, '').replace(/\.html$/, '');
      return clean ? `: ${clean}` : '';
    },

    // The referral ref without a leading @, empty when absent.
    refName(e) {
      return (e.ref || '').replace(/^@/, '');
    },

    // " (iPhone, Safari) from Berlin, DE (203.0.113.9)" suffix for
    // portfolio events. Private and unknown locations stay hidden, the
    // client IP always goes last in parentheses when the log carried one.
    siteDetail(e) {
      const parts = [e.device, e.browser].filter((v) => v);
      let out = parts.length ? ` (${parts.join(', ')})` : '';
      const loc = this.$utils.prettyLocation(e.location);
      if (loc) {
        out += ` from ${loc}`;
      }
      if (e.ip) {
        out += ` (${e.ip})`;
      }
      return out;
    },

    // 1 -> 1st, 2 -> 2nd, 3 -> 3rd, 4 -> 4th, 11-13 -> th.
    ordinal(n) {
      const mod100 = n % 100;
      if (mod100 >= 11 && mod100 <= 13) {
        return `${n}th`;
      }
      const mod10 = n % 10;
      if (mod10 === 1) {
        return `${n}st`;
      }
      if (mod10 === 2) {
        return `${n}nd`;
      }
      if (mod10 === 3) {
        return `${n}rd`;
      }
      return `${n}th`;
    },

    eventTime(stamp) {
      return stamp ? dayjs(stamp).format('DD MMM HH:mm') : '';
    },

    eventRelative(stamp) {
      return stamp ? dayjs(stamp).fromNow() : '';
    },

    manualRefresh() {
      this.fetchData();
    },

    startAutoRefresh() {
      this.stopAutoRefresh();
      this.refreshTimer = setInterval(() => {
        if (document.hidden) {
          return;
        }
        this.fetchData(true);
      }, 60000);
    },

    stopAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer);
        this.refreshTimer = null;
      }
    },

    onVisibilityChange() {
      if (!document.hidden) {
        this.fetchData(true);
      }
    },
  },

  computed: {
    ...mapState(['settings']),
    dayjs() {
      return dayjs;
    },
    // Rows still in the feed, ie everything Simon has not hidden.
    activeEvents() {
      return this.events.filter((e) => !this.hiddenEvents[this.eventKey(e)]);
    },
    // Hiding a row lets the next older one up, so the card keeps its length.
    visibleEvents() {
      if (this.eventsExpanded) {
        return this.activeEvents;
      }
      return this.activeEvents.slice(0, 10);
    },
  },

  created() {
    this.$root.$on('page.refresh', this.manualRefresh);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.manualRefresh);
    this.stopAutoRefresh();
    document.removeEventListener('visibilitychange', this.onVisibilityChange);
  },

  mounted() {
    this.loadHiddenEvents();
    this.fetchData();
    document.addEventListener('visibilitychange', this.onVisibilityChange);
    this.startAutoRefresh();
  },
});
</script>

<style scoped>
.partial-alert {
  margin-bottom: 1.25rem;
  padding: 0.9rem 1.1rem;
  border-radius: 6px;
  border: 2px solid #7f1d1d;
  border-left-width: 10px;
  background: #b91c1c;
  color: #fff;
  box-shadow: 0 2px 10px rgba(185, 28, 28, 0.35);
}
.partial-alert-head {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  font-size: 1.05rem;
  font-weight: 800;
  letter-spacing: 0.01em;
  text-transform: none;
}
.partial-alert-head .icon {
  color: #fff;
  font-size: 1.4rem;
}
.partial-alert-list {
  margin: 0.6rem 0 0 0;
  list-style: none;
}
.partial-alert-list li {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.5rem;
  padding: 0.3rem 0;
  border-top: 1px solid rgba(255, 255, 255, 0.35);
}
.partial-alert .partial-name {
  color: #fff;
  font-weight: 700;
  text-decoration: underline;
}
.partial-alert .partial-count strong {
  color: #fff;
  font-weight: 800;
  background: rgba(0, 0, 0, 0.28);
  padding: 0 0.35rem;
  border-radius: 3px;
}
.partial-alert .partial-when {
  opacity: 0.85;
  font-size: 0.85rem;
}
.partial-alert .partial-dismiss {
  margin-left: auto;
  color: #fff;
  opacity: 0.9;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.partial-alert .partial-dismiss:hover {
  opacity: 1;
  text-decoration: underline;
}
.partial-alert .partial-help {
  margin: 0.6rem 0 0 0;
  font-size: 0.82rem;
  opacity: 0.92;
}
.events .event {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  padding: 0.3rem 0;
  border-bottom: 1px solid #f0f0f0;
}
.events .event:last-child {
  border-bottom: none;
}
.events .event-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
/* Sits at the very right of the row and only shows while the row is hovered.
   The slot is reserved at all times so the text does not reflow on hover. */
.events .event-del {
  flex: 0 0 auto;
  width: 1.4rem;
  align-self: center;
  padding: 0;
  border: none;
  background: none;
  color: #dc2626;
  line-height: 1;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.12s ease-in;
}
.events .event:hover .event-del,
.events .event-del:focus {
  opacity: 1;
}
.events .event-del:hover {
  color: #b91c1c;
}
.events .event-text a {
  text-decoration: none;
}
.events .event-text a:hover {
  text-decoration: underline;
}
.events .timestamp {
  white-space: nowrap;
}
.toggle {
  display: inline-block;
  margin-top: 0.5rem;
}
</style>
