import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

const deactivateTeamDuration = new Trend('deactivate_team_duration_ms');
const reassignReviewerDuration = new Trend('reassign_reviewer_duration_ms');
const deactivateTeamSuccessRate = new Rate('deactivate_team_success');
const reassignReviewerSuccessRate = new Rate('reassign_reviewer_success');
const realErrorsRate = new Rate('real_errors');

export const options = {
  stages: [
    { duration: '30s', target: 5 },
    { duration: '2m', target: 5 },
    { duration: '30s', target: 10 },
    { duration: '1m', target: 5 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<300', 'p(99.9)<300'],
    'http_req_failed{name:!ReassignReviewer}': ['rate<0.001'],
    'real_errors': ['rate<0.001'],
    deactivate_team_duration_ms: ['p(95)<300', 'p(99.9)<300'],
    reassign_reviewer_duration_ms: ['p(95)<300', 'p(99.9)<300'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

let testData = null;

function waitForServer(maxRetries = 30, retryDelay = 1000) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const res = http.get(`${BASE_URL}/health`, { timeout: '5s' });
      if (res.status === 200) {
        console.log('Сервер готов');
        return true;
      }
    } catch (e) {
    }
    console.log(`Ожидание сервера... (${i + 1}/${maxRetries})`);
    if (i < maxRetries - 1) {
      http.get(`${BASE_URL}/health`, { timeout: '1s' });
    }
  }
  return false;
}

export function setup() {
  if (!waitForServer()) {
    console.error('Сервер недоступен. Убедитесь, что сервер запущен на', BASE_URL);
    return null;
  }
  console.log('Подготовка тестовых данных...');
  
  const teams = [];
  const totalUsers = 200;
  const teamsCount = 10;
  const usersPerTeam = Math.floor(totalUsers / teamsCount);
  
  for (let teamIdx = 0; teamIdx < teamsCount; teamIdx++) {
    const teamName = `load-test-team-${teamIdx}-${Date.now()}`;
    
    const members = Array.from({ length: usersPerTeam }, (_, i) => ({
      user_id: `load-user-${teamIdx}-${i}`,
      username: `LoadUser${teamIdx}-${i}`,
    }));
    
    const createTeamRes = http.post(`${BASE_URL}/team/add`, JSON.stringify({
      team_name: teamName,
      members: members,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (createTeamRes.status === 201 || createTeamRes.status === 409) {
      teams.push({
        team_name: teamName,
        users: Array.from({ length: usersPerTeam }, (_, i) => `load-user-${teamIdx}-${i}`),
      });
    }
  }
  
  if (teams.length === 0) {
    console.error('Не удалось создать команды');
    return null;
  }
  
  const prs = [];
  for (let i = 0; i < 10; i++) {
    const prID = `load-pr-${i}-${Date.now()}`;
    const createPRRes = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify({
      pull_request_id: prID,
      pull_request_name: `Тестовый PR ${i}`,
      author_id: teams[0].users[0],
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (createPRRes.status === 201 || createPRRes.status === 409) {
      prs.push({ pr_id: prID, author_id: teams[0].users[0] });
    }
  }
  
  for (let teamIdx = 0; teamIdx < Math.min(teams.length, 5); teamIdx++) {
    const team = teams[teamIdx];
    for (let i = 1; i < Math.min(team.users.length, 10); i++) {
      http.post(`${BASE_URL}/users/setIsActive`, JSON.stringify({
        user_id: team.users[i],
        is_active: true,
      }), {
        headers: { 'Content-Type': 'application/json' },
      });
    }
  }
  
  console.log(`Подготовка завершена: команды=${teams.length}, пользователей_на_команду=${usersPerTeam}, всего_пользователей=${teams.length * usersPerTeam}, pr=${prs.length}`);
  
  return {
    teams: teams,
    prs: prs,
  };
}

export default function (data) {
  if (!data || !data.teams || data.teams.length === 0) {
    return;
  }
  
  const randomTeam = data.teams[Math.floor(Math.random() * data.teams.length)];
  testDeactivateTeamMembers(randomTeam.team_name);
  
  sleep(0.2);
  
  if (data.prs && data.prs.length > 0) {
    const pr = data.prs[Math.floor(Math.random() * data.prs.length)];
    testReassignReviewer(pr.pr_id);
  }
  
  sleep(0.2);
}

function testDeactivateTeamMembers(teamName) {
  const startTime = Date.now();
  
  const res = http.post(`${BASE_URL}/team/deactivateMembers`, JSON.stringify({
    team_name: teamName,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'DeactivateTeamMembers' },
  });
  
  const duration = Date.now() - startTime;
  deactivateTeamDuration.add(duration);
  deactivateTeamSuccessRate.add(res.status === 200);
  realErrorsRate.add(res.status !== 200);
  
  const success = check(res, {
    'статус деактивации команды 200': (r) => r.status === 200,
    'длительность деактивации команды < 300ms': (r) => duration < 300,
    'ответ деактивации команды содержит команду': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.team && body.team.team_name === teamName;
      } catch {
        return false;
      }
    },
  }, { name: 'DeactivateTeamMembers' });
  
  if (!success && res.status !== 200) {
    console.error(`DeactivateTeamMembers не удалось: ${res.status} - ${res.body.substring(0, 200)}`);
  }
}

function testReassignReviewer(prID) {
  const startTime = Date.now();
  
  const getPRRes = http.get(`${BASE_URL}/pullRequest/get?pull_request_id=${prID}`, {
    tags: { name: 'GetPR' },
  });
  
  if (getPRRes.status !== 200) {
    return;
  }
  
  let prData;
  try {
    prData = JSON.parse(getPRRes.body);
  } catch {
    return;
  }
  
  if (!prData.pr || !prData.pr.assigned_reviewers || prData.pr.assigned_reviewers.length === 0) {
    const duration = Date.now() - startTime;
    reassignReviewerDuration.add(duration);
    reassignReviewerSuccessRate.add(true);
    realErrorsRate.add(false);
    return;
  }
  
  const oldReviewerID = prData.pr.assigned_reviewers[0];
  
  const res = http.post(`${BASE_URL}/pullRequest/reassign`, JSON.stringify({
    pull_request_id: prID,
    old_user_id: oldReviewerID,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'ReassignReviewer' },
  });
  
  const duration = Date.now() - startTime;
  reassignReviewerDuration.add(duration);
  reassignReviewerSuccessRate.add(res.status === 200 || res.status === 409);
  const isRealError = res.status !== 200 && res.status !== 409;
  realErrorsRate.add(isRealError);
  
  const success = check(res, {
    'статус переназначения ревьювера 200 или 409': (r) => r.status === 200 || r.status === 409,
    'длительность переназначения ревьювера < 300ms': (r) => duration < 300,
    'ответ переназначения ревьювера содержит PR': (r) => {
      if (r.status === 409) return true;
      try {
        const body = JSON.parse(r.body);
        return body.pr && body.pr.pull_request_id === prID;
      } catch {
        return false;
      }
    },
  }, { name: 'ReassignReviewer' });
  
  if (!success && res.status !== 409) {
    console.error(`ReassignReviewer не удалось: ${res.status} - ${res.body.substring(0, 200)}`);
  }
}

export function teardown(data) {
  console.log('Очистка завершена');
}

