# TypeMUX Documentation

This directory contains the complete documentation for TypeMUX, ready to be deployed as a GitHub Pages site.

## Documentation Structure

- **index.md** - Main landing page with overview and introduction
- **quickstart.md** - 5-minute quick start guide
- **tutorial.md** - Comprehensive step-by-step tutorial
- **reference.md** - Complete language reference and syntax specification
- **examples.md** - Real-world examples demonstrating features
- **configuration.md** - CLI flags, annotations, and configuration options

## Deploying to GitHub Pages

### Option 1: GitHub Pages from /docs

1. Push this directory to your repository
2. Go to repository Settings → Pages
3. Set source to "Deploy from a branch"
4. Select branch (e.g., `main`) and folder `/docs/github-site`
5. Click Save

Your site will be available at `https://yourusername.github.io/typemux/`

### Option 2: Custom GitHub Pages Setup

1. Copy contents of this directory to your docs root or gh-pages branch
2. Configure GitHub Pages in repository settings
3. Your site will be published automatically

## Local Preview

To preview the documentation locally with Jekyll:

```bash
# Install dependencies
gem install bundler jekyll

# Create Gemfile (if not exists)
cat > Gemfile << EOF
source 'https://rubygems.org'
gem 'github-pages', group: :jekyll_plugins
EOF

# Install gems
bundle install

# Serve locally
bundle exec jekyll serve

# Visit http://localhost:4000
```

## Customization

### Update Repository Information

Edit `_config.yml` and update:
- `repository`: Your GitHub username/repo
- `title`: Site title
- `description`: Site description

### Theme Customization

The site uses the Cayman theme by default. To change:

1. Edit `_config.yml`
2. Change `theme: jekyll-theme-cayman` to another supported theme:
   - `jekyll-theme-minimal`
   - `jekyll-theme-architect`
   - `jekyll-theme-slate`
   - `jekyll-theme-modernist`
   - etc.

See [GitHub Pages themes](https://pages.github.com/themes/) for options.

### Custom CSS

Create `assets/css/style.scss`:

```scss
---
---

@import "{{ site.theme }}";

// Your custom CSS here
```

### Custom Navigation

Edit the navigation section in `_config.yml`.

## Content Updates

All documentation is written in Markdown with GitHub Flavored Markdown (GFM) support.

### Adding New Pages

1. Create `newpage.md` in this directory
2. Add front matter:
```yaml
---
title: New Page Title
---
```
3. Add link in navigation (`_config.yml`)

### Syntax Highlighting

Code blocks use Rouge syntax highlighting:

\`\`\`typemux
type User {
  id: string @required
}
\`\`\`

Supported languages:
- `typemux` (will be highlighted as similar language)
- `graphql`
- `protobuf`
- `yaml`
- `bash`
- `json`

## Maintenance

### Keeping Documentation in Sync

When updating the TypeMUX codebase:
1. Update relevant documentation pages
2. Add new examples if features are added
3. Update the reference if syntax changes
4. Add migration notes if breaking changes occur

### Version Documentation

For version-specific docs, consider:
- Creating subdirectories per version (v1, v2, etc.)
- Using git tags/branches for versioned documentation
- Adding version selector in custom navigation

## Contributing

When contributing to documentation:
1. Ensure all code examples are tested
2. Keep writing style consistent
3. Add cross-references between related sections
4. Update table of contents if adding sections
5. Preview locally before submitting PR

## Links

- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [Jekyll Documentation](https://jekyllrb.com/docs/)
- [Markdown Guide](https://www.markdownguide.org/)
