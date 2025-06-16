import http from 'k6/http';
import { check } from 'k6';
import { randomItem, uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    vus: 10,           // number of virtual users
    duration: '30s',   // total test duration
};

const sampleTitles = [
    "Advance Wars: Tactical Turn-Based Combat",
    "Fire Emblem: Strategic Warfare",
    "Final Fantasy Tactics: Legacy of Ivalice",
    "Into the Breach: Minimalist Mayhem",
    "Wargroove: Modern Retro Strategy"
];

const sampleContents = [
    "# Advance Wars\nReleased in 2001 for the Game Boy Advance, *Advance Wars* is a turn-based strategy game...",
    "# Fire Emblem\nA medieval strategy RPG known for its permadeath and deep storytelling...",
    "# Final Fantasy Tactics\nSet in the world of Ivalice, this game revolutionized the genre...",
    "# Into the Breach\nA minimalist tactical game involving time travel and giant insects...",
    "# Wargroove\nA modern love letter to Advance Wars with new commanders and multiplayer..."
];

export default function () {
    const url = 'http://localhost:3000/api/v1/note';
    const payload = JSON.stringify({
        title: `${randomItem(sampleTitles)} (${uuidv4()})`,
        content: randomItem(sampleContents),
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const res = http.post(url, payload, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'note created': (r) => r.body && r.body.includes("title"),
    });
}