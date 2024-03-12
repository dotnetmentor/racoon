<script>
import { handleError } from 'vue'

const pageArray = (array, size) => {
  const chunkedArr = []
  let index = 0
  while (index < array.length) {
    chunkedArr.push(array.slice(index, size + index))
    index += size
  }
  return chunkedArr
}

export default {
  data() {
    return {
      loading: false,
      error: null,
      filters: {},
      configs: [],
      pages: [],
      matching: 0,
      searchFilters: [],
      searchChanged: false,
      total: 0,
      more: false,
      property: null,
    }
  },

  methods: {
    async queryConfigs(download, startAt) {
      try {
        this.loading = true
        this.error = null
        if (download) {
          this.matching = 0
          this.total = 0
          this.more = false
        }

        let url = `/api/query/config`
        let params = []

        if (download) {
          params.push('download=true')
        }

        if (startAt) {
          params.push(`startAt=${startAt}`)
        }

        this.activeFilters.forEach(f => {
          params.push(`f=${f.key}/${f.value}`)
        })

        if (params.length > 0) {
          url += `?${params.join('&')}`
        }

        console.log('fetching data', url)
        const response = await fetch(url, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          }
        })

        const data = await response.json()
        if (data.error) {
          throw data.error
        }

        this.handleQueryResponse(data, download, startAt > 0)
      } catch (e) {
        console.error(e)
        this.error = e
      } finally {
        this.loading = false
      }
    },

    async decryptConfig(config, index) {
      try {
        if (!config.encrypted) {
          return
        }
        config.decrypting = true

        console.log('decrypting config', config)
        this.error = null

        let url = `/api/command/config/decrypt`
        let request = {
          path: config.path,
        }
        console.log('fetching data', url, request)
        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(request),
        })

        const data = await response.json()
        console.log('handling decryption response', data)

        if (data.error) {
          throw data.error
        }

        console.log('decrypted config', config)
        config.properties = data.data.properties
        config.encrypted = false
        this.configs[index] = config
      } catch (e) {
        console.error(e)
        this.error = e
      } finally {
        config.decrypting = false
      }
    },

    handleQueryResponse(data, download, append) {
      console.log('handeling query response', data)

      let allFilters = []
      let allConfigs = []

      data.items.forEach(x => {
        let parts = x.path.split('/')
        let filters = []

        for (let i = 0; i < parts.length; i++) {
          if (i >= (parts.length - 2)) {
            let kvp = {
              key: 'name',
              value: parts[i]
            }
            filters.push(kvp)
            break
          }

          filters.push({
            key: parts[i],
            value: parts[i + 1]
          })

          i++
        }

        allFilters.push(...filters)

        let config = {
          ...x.data,
          path: x.path,
          encrypted: x.encrypted,
          filters: {},
        }
        filters.forEach(x => {
          config.filters[x.key] = x.value
        })
        allConfigs.push(config)
      })

      let uniqueFilters = []
      let filters = {}

      allFilters.filter((item, index, self) => {
        let key = `${item.key}=${item.value}`
        if (!uniqueFilters.includes(key)) {
          uniqueFilters.push(key)
          return true
        } else {
          return false
        }
      }).sort((a, b) => {
        if (a.key < b.key) {
          return -1
        }
        if (a.key > b.key) {
          return 1
        }
        return 0
      }).forEach(x => {
        if (!filters[x.key]) {
          filters[x.key] = []
        }
        filters[x.key].push({
          value: x.value
        })
      })

      Object.keys(filters).forEach(key => {
        if (!this.filters[key]) {
          this.filters[key] = filters[key]
        } else {
          filters[key].forEach(f => {
            if (this.filters[key].map(x => x.value).includes(f.value)) {
              return
            }
            this.filters[key].push(f)
          })
        }
      })

      if (download) {
        if (append) {
          this.configs.push(...allConfigs)
        } else {
          this.configs = allConfigs
        }
        this.total = data.total
        this.more = data.more
        this.searchFilters = data.filters
        this.filterConfigs()
        this.refreshFilters()
      }
    },

    refreshFilters() {
      let changed = false
      this.activeFilters.forEach(f => {
        if (!this.searchFilters.includes(`${f.key}=${f.value}`)) {
          changed = true
        }
      })
      if (!changed && this.searchFilters.length !== this.activeFilters.length) {
        changed = true
      }
      console.log('search changed', changed)
      this.searchChanged = changed
    },

    toggleFilter(filter, key) {
      this.filters[key].forEach(x => {
        if (x.value === filter.value) {
          x.active = !x.active
        }
      })

      this.refreshFilters()
      this.filterConfigs()
    },

    toggleProperty(p, encrypted) {
      if (p) {
        p.encrypted = encrypted
      }
      this.property = p
    },

    filterConfigs() {
      // Must match 1 per filter group (context, name etc)
      let filtered = this.configs.filter(c => {
        let matchesKey = {}
        this.allFilters.forEach(fg => {
          let key = fg[0].key
          let hasActive = fg.some(x => x.active)
          if (!hasActive) {
            return
          }

          let match = false
          let matchValue
          fg.filter(kv => kv.active).forEach(kv => {
            if (!match && c.filters[kv.key] === kv.value) {
              match = true
              matchValue = kv.value
            }
          })

          matchesKey[key] = match
        })
        let matchesFilters = Object.keys(matchesKey).every(key => matchesKey[key])
        console.log('matching all active filter groups', matchesKey, matchesFilters)
        return matchesFilters
      }).sort((a, b) => {
        if (a.matches < b.matches) {
          return -1
        }
        if (a.matches > b.matches) {
          return 1
        }
        return 0
      })
      this.matching = filtered.length
      this.pages = pageArray(filtered, 3)
    },
  },

  computed: {
    allFilters() {
      return Object.keys(this.filters).map(key => {
        return this.filters[key].map(f => {
          return {
            key: key,
            value: f.value,
            active: f.active
          }
        })
      })
    },

    activeFilters() {
      return this.allFilters.flat().filter(x => x.active)
    },

    pagedFilters() {
      let pages = pageArray(this.allFilters, 3)
      return pages
    },
  },

  mounted() {
    this.queryConfigs(false)
  }
}
</script>

<template>
  <article class="userinput">
    <h3>Configurations</h3>

    <hgroup>
      <h4>Filters ( {{ activeFilters.length }} )</h4>
      <div class="grid" v-for="page in pagedFilters">
        <div v-for="values in page">
          <span class="tags">
            <strong>{{ values[0].key }}:</strong><br />
            <span v-for="f in values"
              :class="{ 'filter': true, 'active': f.active, 'search': searchFilters.includes(`${f.key}=${f.value}`) }">
              <kbd @click="() => toggleFilter(f, f.key)">{{ f.value }}</kbd>
            </span>
          </span>
        </div>
      </div>
    </hgroup>

    <hgroup v-if="error">
      <h4>Error</h4>
      <pre><code>{{ error }}</code></pre>
    </hgroup>

    <hgroup v-if="activeFilters.length === 0">
      <h4>Note</h4>
      <p>You must select at least 1 filter!</p>
    </hgroup>

    <div v-if="!loading && activeFilters.length" class="grid">
      <p>
        <mark><strong>{{ configs.length }}</strong> of <strong>{{ total }}</strong></mark>
        configuration(s) loaded from the server.<br />
        <mark><strong>{{ matching }}</strong></mark> matching at least 1 of the selected filter(s).
      </p>
      <div>&nbsp;</div>
      <div>
        <button class="secondary" v-if="searchChanged === true" @click="() => queryConfigs(activeFilters.length, 0)">
          Search
        </button>
        <button class="secondary" v-if="more && searchChanged === false"
          @click="() => queryConfigs(activeFilters.length, configs.length)">
          Load more
        </button>
      </div>
    </div>
  </article>

  <div>
    <div v-if="activeFilters.length && pages.length" v-for="page in pages">
      <div class="config-grid grid">
        <article v-for="c, index in page">
          <header>
            <hgroup>
              <h5>
                {{ c.name }}
                <a v-if="c.encrypted" class="decrypt" @click="() => decryptConfig(c, index)">
                  <v-icon name="fa-unlock-alt" scale="1" :animation="c.decrypting ? 'float' : 'none'" title="Decrypt" />
                </a>
              </h5>
              <div class="tags">
                <template v-for="v, k in c.filters">
                  <span v-if:="k !== 'name'" class="filter">
                    <kbd>{{ k }} = {{ v }}</kbd>
                  </span>
                </template>
              </div>
              <div class="tags">
                <span v-for="v, k in c.labels" class="label">
                  <kbd>{{ k }} = {{ v }}</kbd>
                </span>
              </div>
            </hgroup>
          </header>
          <details open>
            <summary>Properties</summary>
            <figure>
              <table role="grid">
                <thead>
                  <tr>
                    <th scope="col">#</th>
                    <th scope="col" colspan="2">Property</th>
                    <th scope="col">Value</th>
                  </tr>
                </thead>
                <tbody>
                  <template v-for="p, i in c.properties">
                    <tr @click="() => toggleProperty(p, c.encrypted)">
                      <th scope="row">{{ i + 1 }}</th>
                      <td>{{ p.name }}</td>
                      <td><v-icon name="fa-shield-alt" size="1" title="Sensitive" v-if="p.sensitive" /></td>
                      <td>
                        <b v-if="p.sensitive && c.encrypted">{{ '<sensitive>' }}</b>
                        <span v-else>{{ p.value }}</span>
                      </td>
                    </tr>
                  </template>
                </tbody>
              </table>
            </figure>
          </details>
        </article>
      </div>
    </div>

    <div v-if="!loading && activeFilters.length">
      <button class="secondary" v-if="more && searchChanged === false"
        @click="() => queryConfigs(activeFilters.length, configs.length)">
        Load more
      </button>
    </div>

    <hgroup v-if="loading">
      <h4>Loading...</h4>
    </hgroup>

    <dialog :open="property" v-if="property">
      <article>
        <header>
          <hgroup>
            <a aria-label="Close" class="close" @click="() => toggleProperty()"></a>
            <h5>{{ property.name }}</h5>
            <div>{{ property.description }}</div>
          </hgroup>
        </header>
        <div>
          <p v-if="property.sensitive">
            <v-icon name="fa-shield-alt" size="1" title="Sensitive" />
            Property holds sensitive data, treat decrypted values with care!<br /><br />
          </p>
          <pre><code>{{ property.sensitive && property.encrypted ? '<sensitive>' : property.value }}</code></pre>
        </div>
      </article>
    </dialog>

  </div>
</template>
