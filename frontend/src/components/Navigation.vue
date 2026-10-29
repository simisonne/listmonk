<template>
  <b-menu-list>
    <b-menu-item :to="{ name: 'dashboard' }" tag="router-link" :active="activeItem.dashboard"
      icon="view-dashboard-variant-outline" :label="$t('menu.dashboard')" /><!-- dashboard -->

    <b-menu-item :expanded="activeGroup.lists" :active="activeGroup.lists" data-cy="lists"
      @update:active="(state) => toggleGroup('lists', state)" icon="format-list-bulleted-square"
      :label="$t('globals.terms.lists')">
      <b-menu-item :to="{ name: 'lists' }" tag="router-link" :active="activeItem.lists" data-cy="all-lists"
        icon="format-list-bulleted-square" :label="$t('menu.allLists')" />
      <b-menu-item :to="{ name: 'forms' }" tag="router-link" :active="activeItem.forms" class="forms"
        icon="newspaper-variant-outline" :label="$t('menu.forms')" />
    </b-menu-item><!-- lists -->

    <b-menu-item v-if="$can('subscribers:*')" :expanded="activeGroup.subscribers" :active="activeGroup.subscribers"
      data-cy="subscribers" @update:active="(state) => toggleGroup('subscribers', state)" icon="account-multiple"
      :label="$t('globals.terms.subscribers')">
      <b-menu-item v-if="$can('subscribers:get_all', 'subscribers:get')" :to="{ name: 'subscribers' }" tag="router-link"
        :active="activeItem.subscribers" data-cy="all-subscribers" icon="account-multiple"
        :label="$t('menu.allSubscribers')" />
      <b-menu-item v-if="$can('subscribers:import')" :to="{ name: 'import' }" tag="router-link"
        :active="activeItem.import" data-cy="import" icon="file-upload-outline" :label="$t('menu.import')" />
      <b-menu-item v-if="$can('bounces:get')" :to="{ name: 'bounces' }" tag="router-link" :active="activeItem.bounces"
        data-cy="bounces" icon="email-bounce" :label="$t('globals.terms.bounces')" />
    </b-menu-item><!-- subscribers -->

    <b-menu-item v-if="$can('campaigns:*')" :expanded="activeGroup.campaigns" :active="activeGroup.campaigns"
      data-cy="campaigns" @update:active="(state) => toggleGroup('campaigns', state)" icon="rocket-launch-outline"
      :label="$t('globals.terms.campaigns')">
      <b-menu-item v-if="$can('campaigns:get')" :to="{ name: 'campaigns' }" tag="router-link"
        :active="activeItem.campaigns" data-cy="all-campaigns" icon="rocket-launch-outline"
        :label="$t('menu.allCampaigns')" />
      <b-menu-item v-if="$can('campaigns:manage')" :to="{ name: 'campaign', params: { id: 'new' } }" tag="router-link"
        :active="activeItem.campaign" data-cy="new-campaign" icon="plus" :label="$t('menu.newCampaign')" />
      <b-menu-item v-if="$can('media:*')" :to="{ name: 'media' }" tag="router-link" :active="activeItem.media"
        data-cy="media" icon="image-outline" :label="$t('menu.media')" />
      <b-menu-item v-if="$can('templates:get')" :to="{ name: 'templates' }" tag="router-link"
        :active="activeItem.templates" data-cy="templates" icon="file-image-outline"
        :label="$t('globals.terms.templates')" />
      <b-menu-item v-if="$can('campaigns:get_analytics')" :to="{ name: 'campaignAnalytics' }" tag="router-link"
        :active="activeItem.campaignAnalytics" data-cy="analytics" icon="chart-bar"
        :label="$t('globals.terms.analytics')" />
    </b-menu-item><!-- campaigns -->

    <b-menu-item v-if="$can('users:*', 'roles:*')" :expanded="activeGroup.users" :active="activeGroup.users"
      data-cy="users" @update:active="(state) => toggleGroup('users', state)" icon="account-multiple"
      :label="$t('globals.terms.users')">
      <b-menu-item v-if="$can('users:get')" :to="{ name: 'users' }" tag="router-link" :active="activeItem.users"
        data-cy="users" icon="account-multiple" :label="$t('globals.terms.users')" />
      <b-menu-item v-if="$can('roles:get')" :to="{ name: 'userRoles' }" tag="router-link" :active="activeItem.userRoles"
        data-cy="userRoles" icon="newspaper-variant-outline" :label="$t('users.userRoles')" />
      <b-menu-item v-if="$can('roles:get')" :to="{ name: 'listRoles' }" tag="router-link" :active="activeItem.listRoles"
        data-cy="listRoles" icon="format-list-bulleted-square" :label="$t('users.listRoles')" />
    </b-menu-item><!-- users -->

    <!-- Website: every tracked site event, between Users and Settings. The
         icon font here has no globe glyph, so the globe is inline SVG. -->
    <b-menu-item :to="{ name: 'website' }" tag="router-link" :active="activeItem.website"
      data-cy="website">
      <template #label>
        <span class="nav-website-label">
          <svg class="nav-globe" viewBox="0 0 24 24" aria-hidden="true">
            <path :d="globePath" />
          </svg>
          {{ $t('menu.website') }}
        </span>
      </template>
    </b-menu-item><!-- website -->

    <b-menu-item v-if="$can('settings:*')" :expanded="activeGroup.settings" :active="activeGroup.settings"
      data-cy="settings" @update:active="(state) => toggleGroup('settings', state)" icon="cog-outline"
      :label="$t('menu.settings')">
      <b-menu-item v-if="$can('settings:get')" :to="{ name: 'settings' }" tag="router-link"
        :active="activeItem.settings" data-cy="all-settings" icon="cog-outline" :label="$t('menu.settings')" />
      <b-menu-item v-if="$can('settings:maintain')" :to="{ name: 'maintenance' }" tag="router-link"
        :active="activeItem.maintenance" data-cy="maintenance" icon="wrench-outline" :label="$t('menu.maintenance')" />
      <b-menu-item v-if="$can('settings:get')" :to="{ name: 'logs' }" tag="router-link" :active="activeItem.logs"
        data-cy="logs" icon="format-list-bulleted-square" :label="$t('menu.logs')" />
    </b-menu-item><!-- settings -->
  </b-menu-list>
</template>

<script>
import { mapState } from 'vuex';

export default {
  name: 'Navigation',

  props: {
    activeItem: { type: Object, default: () => { } },
    activeGroup: { type: Object, default: () => { } },
    isMobile: Boolean,
  },

  data() {
    return {
      // Material's globe outline. The bundled icon font has no globe glyph,
      // so the one menu icon that needs it draws its own path. Split at path
      // command boundaries only, the pieces concatenate into one path.
      globePath: [
        'M17.9 17.39c-.26-.8-1.01-1.39-1.9-1.39h-1v-3a1 1 0 0 0-1-1H8v-2h2',
        'a1 1 0 0 0 1-1V7h2a2 2 0 0 0 2-2v-.41a7.984 7.984 0 0 1 2.9 12.8',
        'M11 19.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1a2 2 0 0 0 2 2',
        'm1-16A10 10 0 0 0 2 12a10 10 0 0 0 10 10 10 10 0 0 0 10-10A10 10 0 0 0 12 2Z',
      ].join(''),
    };
  },

  methods: {
    toggleGroup(group, state) {
      this.$emit('toggleGroup', group, state);
    },

    doLogout() {
      this.$emit('doLogout');
    },
  },

  computed: {
    ...mapState(['profile']),
  },

  mounted() {
    // A hack to close the open accordion burger menu items on click.
    // Buefy does not have a way to do this.
    if (this.isMobile) {
      document.querySelectorAll('.navbar li a[href]').forEach((e) => {
        e.onclick = () => {
          document.querySelector('.navbar-burger').click();
        };
      });
    }
  },
};

</script>

<style scoped>
/* The globe sits where a b-icon sits in every other menu item. */
.nav-website-label {
  display: inline-flex;
  align-items: center;
  gap: 0.5em;
}
.nav-globe {
  width: 1.05em;
  height: 1.05em;
  flex: none;
  fill: currentColor;
}
</style>
