import { cp, mkdir, readdir, readFile, rename, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { execSync } from 'node:child_process';
import process from 'node:process';
import { fileURLToPath } from 'node:url';
import { createHash } from 'node:crypto';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const appDir = path.resolve(scriptDir, '..');
const staticNextDir = path.resolve(appDir, '../../../server/internal/webapp/static-next');
const staticNextBackupDir = path.resolve(appDir, '../../../server/internal/webapp/static-next.__backup');
const keepFiles = new Set(['.gitignore', 'README.md']);
const staticNextGitignore = `*
!.gitignore
!README.md
`;
const staticNextReadme = `# Svelte Root Build Output

- Source app: \`apps/web/svelte\`
- Publish command: \`npm --prefix apps/web/svelte run publish:go-static\`
- Root smoke command: \`npm --prefix apps/web/svelte run smoke:root\`
- Note: restart/rebuild the Go server after publishing so embedded assets are refreshed.
`;

const requiredSmokeRoutes = [
	'/',
	'/signin',
	'/setup',
	'/movies',
	'/tv',
	'/settings',
	'/admin',
	'/movies/[id]',
	'/tv/[id]'
];

async function cleanStaticNextDirectory() {
	await mkdir(staticNextDir, { recursive: true });
	const entries = await readdir(staticNextDir, { withFileTypes: true });
	await Promise.all(
		entries
			.filter((entry) => !keepFiles.has(entry.name))
			.map((entry) =>
				rm(path.join(staticNextDir, entry.name), {
					force: true,
					recursive: true
				})
			)
	);
}

async function snapshotExistingStaticNext() {
	await rm(staticNextBackupDir, { recursive: true, force: true });
	try {
		await rename(staticNextDir, staticNextBackupDir);
	} catch (error) {
		const code = String(error?.code || '').toUpperCase();
		if (code === 'ENOENT') return;
		if (code !== 'EXDEV') throw error;
		await cp(staticNextDir, staticNextBackupDir, { recursive: true, force: true });
		await rm(staticNextDir, { recursive: true, force: true });
	}
}

async function restoreStaticNextSnapshot() {
	await rm(staticNextDir, { recursive: true, force: true });
	try {
		await rename(staticNextBackupDir, staticNextDir);
	} catch (error) {
		const code = String(error?.code || '').toUpperCase();
		if (code === 'ENOENT') return;
		if (code !== 'EXDEV') throw error;
		await cp(staticNextBackupDir, staticNextDir, { recursive: true, force: true });
		await rm(staticNextBackupDir, { recursive: true, force: true });
	}
}

async function discardStaticNextSnapshot() {
	await rm(staticNextBackupDir, { recursive: true, force: true });
}

const textAssetExtensions = new Set(['.css', '.html', '.js', '.json', '.md', '.svg', '.txt']);

async function normalizeStaticNextTextFiles(dir = staticNextDir) {
	const entries = await readdir(dir, { withFileTypes: true });
	for (const entry of entries) {
		const abs = path.join(dir, entry.name);
		if (entry.isDirectory()) {
			await normalizeStaticNextTextFiles(abs);
			continue;
		}
		if (!entry.isFile()) continue;
		if (!textAssetExtensions.has(path.extname(entry.name).toLowerCase())) continue;

		const source = await readFile(abs, 'utf8');
		const normalized = source.replace(/[ \t]+$/gm, '');
		if (normalized !== source) {
			await writeFile(abs, normalized, 'utf8');
		}
	}
}

function runNpmBuild() {
	execSync('npm run build', {
		cwd: appDir,
		stdio: 'inherit',
		shell: true
	});
}

function readGitCommit() {
	try {
		return execSync('git rev-parse --short HEAD', {
			cwd: appDir,
			stdio: ['ignore', 'pipe', 'ignore'],
			shell: true
		})
			.toString('utf8')
			.trim();
	} catch {
		return '';
	}
}

async function discoverRoutePatterns() {
	const routesRoot = path.resolve(appDir, 'src/routes');
	const files = [];

	async function walk(dir, relative = '') {
		const entries = await readdir(dir, { withFileTypes: true });
		for (const entry of entries) {
			const abs = path.join(dir, entry.name);
			const rel = relative ? path.join(relative, entry.name) : entry.name;
			if (entry.isDirectory()) {
				await walk(abs, rel);
				continue;
			}
			if (entry.isFile() && entry.name === '+page.svelte') {
				files.push(relative);
			}
		}
	}

	await walk(routesRoot);

	return files
		.map((relativeDir) => {
			if (!relativeDir) return '/';
			const urlPath = relativeDir
				.replaceAll('\\', '/')
				.split('/')
				.filter(Boolean)
				.join('/');
			return `/${urlPath}`;
		})
		.sort((a, b) => a.localeCompare(b));
}

async function writeBuildInfo() {
	const indexPath = path.join(staticNextDir, 'index.html');
	const indexContent = await readFile(indexPath, 'utf8');
	const indexHash = createHash('sha256').update(indexContent).digest('hex').slice(0, 16);
	const routePatterns = await discoverRoutePatterns();
	const gitCommit = readGitCommit();
	const publishedAt = new Date().toISOString();
	const buildID = `${publishedAt.replace(/[-:.TZ]/g, '').slice(0, 14)}-${indexHash}`;

	const payload = {
		buildID,
		publishedAt,
		basePath: '/',
		sourceApp: 'apps/web/svelte',
		staticOutput: 'server/internal/webapp/static-next',
		gitCommit: gitCommit || null,
		indexHash,
		routePatterns,
		requiredSmokeRoutes
	};

	await writeFile(path.join(staticNextDir, 'build-info.json'), `${JSON.stringify(payload, null, 2)}\n`, 'utf8');
}

async function ensureManagedFiles() {
	await writeFile(path.join(staticNextDir, '.gitignore'), staticNextGitignore, 'utf8');
	await writeFile(path.join(staticNextDir, 'README.md'), staticNextReadme, 'utf8');
}

await snapshotExistingStaticNext();
try {
	await cleanStaticNextDirectory();
	runNpmBuild();
	await normalizeStaticNextTextFiles();
	await writeBuildInfo();
	await ensureManagedFiles();
	await discardStaticNextSnapshot();
} catch (error) {
	await restoreStaticNextSnapshot();
	throw error;
}
