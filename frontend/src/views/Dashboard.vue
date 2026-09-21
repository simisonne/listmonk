<template>
  <section class="dashboard content">
    <header class="columns">
      <div class="column is-two-thirds">
        <h1 class="title is-5">
          {{ $utils.niceDate(new Date()) }}
        </h1>
      </div>
    </header>

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
                <div class="columns" data-cy="messages">
                  <div class="column is-12">
                    <p class="title">
                      <b-icon icon="email-outline" />
                      {{ $utils.niceNumber(counts.messages) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $t('dashboard.messagesSent') }}
                    </p>
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
                <li v-for="(e, i) in visibleEvents" :key="i" class="event">
                  <b-icon :icon="eventIcon(e.type)" size="is-small" />
                  <span class="event-text">{{ eventText(e) }}</span>
                  <small class="has-text-grey" :title="eventRelative(e.created_at)">{{ eventTime(e.created_at) }}</small>
                </li>
              </ul>
              <p v-else-if="!isEventsLoading" class="has-text-grey">
                {{ $t('dashboard.noEvents') }}
              </p>
              <a v-if="events.length > 5" href="#" class="toggle"
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
import Chart from '../components/Chart.vue';

dayjs.extend(relativeTime);

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
      campaignViews: null,
      campaignClicks: null,
      counts: {
        lists: {},
        subscribers: {},
        campaigns: {},
        messages: 0,
      },
    };
  },

  methods: {
    fetchData() {
      this.isCountsLoading = true;
      this.isChartsLoading = true;

      this.$api.getDashboardCounts().then((data) => {
        this.counts = data;
        this.isCountsLoading = false;
      });

      this.$api.getDashboardCharts().then((data) => {
        this.isChartsLoading = false;
        this.campaignViews = this.makeChart(data.campaignViews);
        this.campaignClicks = this.makeChart(data.linkClicks);
      });

      // Fetch the 20 newest events once; the card shows 5 until expanded.
      this.isEventsLoading = true;
      this.$api.getDashboardEvents(20).then((data) => {
        this.events = data;
        this.isEventsLoading = false;
      });
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
        site_visit: 'web',
        track_played: 'music',
        track_downloaded: 'download',
      }[type] || 'bell-outline';
    },

    eventText(e) {
      const who = e.email || e.subscriber_name || 'Someone';
      switch (e.type) {
        case 'campaign_sent':
          return `Campaign sent: ${e.campaign_name || `#${e.campaign_id}`}`;
        case 'open':
          return `${who} opened ${e.campaign_name || 'a campaign'}`;
        case 'click':
          return `${who} clicked ${this.linkHost(e.url)}in ${e.campaign_name || 'a campaign'}`;
        case 'optin':
          return `${who} joined ${e.list_name || 'a public list'}`;
        case 'site_visit':
          return 'Melodies site visited';
        case 'track_played':
          return `Melodies track played${e.track ? `: ${e.track}` : ''}`;
        case 'track_downloaded':
          return `Melodies track downloaded${e.track ? `: ${e.track}` : ''}`;
        default:
          return e.type;
      }
    },

    eventTime(stamp) {
      return stamp ? dayjs(stamp).format('DD MMM HH:mm') : '';
    },

    eventRelative(stamp) {
      return stamp ? dayjs(stamp).fromNow() : '';
    },

    linkHost(url) {
      if (!url) {
        return 'a link ';
      }
      try {
        return `${new URL(url).hostname} `;
      } catch (err) {
        return 'a link ';
      }
    },
  },

  computed: {
    ...mapState(['settings']),
    dayjs() {
      return dayjs;
    },
    visibleEvents() {
      if (this.eventsExpanded) {
        return this.events;
      }
      return this.events.slice(0, 5);
    },
  },

  created() {
    this.$root.$on('page.refresh', this.fetchData);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.fetchData);
  },

  mounted() {
    this.fetchData();
  },
});
</script>

<style scoped>
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
.toggle {
  display: inline-block;
  margin-top: 0.5rem;
}
</style>
