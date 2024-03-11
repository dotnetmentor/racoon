<template>
  <aside>
    <h1>
      <v-icon name="fa-shield-alt" scale="3" fill="#02151A" animation="pulse" speed="slow" /> racoon
    </h1>
    <nav>
      <ul>
        <li><router-link to="/"><v-icon name="fa-home" scale="1.2" /> Start</router-link></li>
        <li><router-link to="/configs"><v-icon name="fa-cogs" scale="1.2" /> Configurations</router-link></li>
        <li><router-link to="/compare"><v-icon name="fa-arrows-alt-h" scale="1.2" /> Compare</router-link></li>
      </ul>
      <ul class="footer">
        <li><a href="https://github.com/dotnetmentor/racoon" data-tooltip="Github"><v-icon name="fa-github"
              scale="3" /></a></li>
      </ul>
    </nav>
  </aside>

  <main class="container-fluid">
    <router-view />
  </main>
</template>
