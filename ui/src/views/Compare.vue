<script>
export default {
  components: {
  },
  data() {
    return {
      error: null,
      loading: false,
      modes: ['unified', 'split'],
      mode: 'split',
      source: {
        left: 'export',
        right: 'export'
      },
      sources: ['text', 'read', 'export'],
      textinput: {
        left: '',
        right: '',
      },
      args: {
        left: '-p context=local -o dotenv',
        right: '-p context=dev -o dotenv',
      },
      results: {
        left: {
          result: 'This is the left side',
          logs: 'No logs available...'
        },
        right: {
          result: 'This is the right side',
          logs: 'No logs available...'
        },
      },
    }
  },
  methods: {
    sourceChanged(side) {
      if (this.source.left === 'text' || this.source.right === 'text') {
        return
      }
      if (side === 'left' && this.source.left !== 'text') {
        this.source.right = this.source.left
      }
      if (side === 'right' && this.source.right !== 'text') {
        this.source.left = this.source.right
      }
    },
    async compare() {
      try {
        this.loading = true
        let request = {}

        if (this.source.left == 'text') {
          this.results.left.result = this.textinput.left
          this.results.left.logs = ''
        } else {
          request.command = this.source.left
          request.left = this.args.left
        }

        if (this.source.right == 'text') {
          this.results.right.result = this.textinput.right
          this.results.right.logs = ''
        } else {
          request.command = this.source.right
          request.right = this.args.right
        }

        if (this.source.left == 'text' && this.source.right == 'text') {
          return
        }

        let url = `/api/command/compare`
        console.log('fetching compare data', request)
        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(request),
        })
        const data = await response.json()
        console.log('recieved compare data', data)
        if (data.left) {
          this.results.left = data.left
        }
        if (data.right) {
          this.results.right = data.right
        }
      } catch (e) {
        this.error = e
        this.loading = false
      } finally {
        this.loading = false
      }
    }
  },
}
</script>

<template>
  <section>
    <h3>Compare</h3>

    <div class="grid">
      <div>
        <label>Left Source<br />
          <select v-model="source.left" id="left-source" @change="() => sourceChanged('left')">
            <option :key="val" v-for="val in sources">{{ val }}</option>
          </select>
        </label>
      </div>
      <div>
        <label>Right Source<br />
          <select v-model="source.right" id="right-source" @change="() => sourceChanged('right')">
            <option :key="val" v-for="val in sources">{{ val }}</option>
          </select>
        </label>
      </div>
    </div>

    <div class="grid">
      <div v-if="source.left == 'text'">
        <label>Left text<br />
          <textarea v-model="textinput.left" rows="6" cols="50"></textarea>
        </label>
      </div>
      <div v-else>
        <label>Left args<br />
          <input type="text" v-model="args.left" />
        </label>
      </div>

      <div v-if="source.right == 'text'">
        <label>Right text<br />
          <textarea v-model="textinput.right" rows="6" cols="50"></textarea>
        </label>
      </div>
      <div v-else>
        <label>Right args<br />
          <input type="text" v-model="args.right" />
        </label>
      </div>
    </div>

    <div class="grid">
      <div>
        <label><br />
          <button :disabled="loading" :aria-busy="loading" @click="() => compare()">Compare now</button>
        </label>
      </div>
    </div>
  </section>

  <section v-if="!loading">
    <div class="grid">
      <div>
        <h3>Diff</h3>
      </div>
      <div>
        <select v-model="mode" id="mode">
          <option :key="val" v-for="val in modes">{{ val }}</option>
        </select>
      </div>
    </div>

    <div class="grid">
      <Diff :mode="mode" theme="light" language="html" :prev="results.left.result" :current="results.right.result" />
    </div>
  </section>

  <section v-if="!loading">
    <div class="grid">
      <div>
        <h4>Logs</h4>
      </div>
    </div>
    <div class="grid">
      <Diff mode="split" theme="light" language="html" :prev="results.left.logs" :current="results.right.logs" />
    </div>
  </section>
</template>
