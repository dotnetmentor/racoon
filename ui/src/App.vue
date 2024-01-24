<template>
  <header class="container-fluid">
    <nav>
      <ul>
        <li><router-link to="/">Start</router-link></li>
        <li><router-link to="/configs">Configurations</router-link></li>
        <li><router-link to="/compare">Compare</router-link></li>
      </ul>
      <ul>
        <li>
          <h1><v-icon name="fa-shield-alt" scale="3" fill="#28284a" animation="pulse" speed="slow" /> Racoon</h1>
        </li>
      </ul>
    </nav>
  </header>

  <main class="container-fluid">
    <router-view />
  </main>
</template>
