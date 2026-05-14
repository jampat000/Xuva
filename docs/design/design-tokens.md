# Xuva Design Tokens

These tokens define the first visual direction for Xuva. They are intentionally small until the first real UI implementation creates pressure for more.

## Color

```text
background/cinema        #080705
background/ink           #11100D
background/panel         #191713
background/elevated      #211E19

text/primary             #FFF7EA
text/secondary           #B9B0A2
text/muted               #746C60
text/inverse             #080705

border/subtle            #312B23
border/strong            #5A4C3D

accent/focus             #4FE6C8
accent/action            #F6B756
accent/live              #FF5D5D
accent/success           #65D887
accent/info              #7CA7FF

media/scrim              rgba(8, 7, 5, 0.72)
media/scrim-strong       rgba(8, 7, 5, 0.88)
```

## Typography

Primary UI font target:

- Inter, SF Pro, or system UI depending on platform.

TV scale:

```text
display/hero             64px / 68px / 700
heading/xl               38px / 44px / 700
heading/lg               28px / 34px / 650
body/lg                  22px / 30px / 450
body/md                  18px / 26px / 450
body/sm                  15px / 22px / 500
caption                  13px / 18px / 600
```

Admin/web scale:

```text
heading/xl               32px / 40px / 700
heading/lg               24px / 32px / 650
heading/md               18px / 26px / 650
body/md                  15px / 24px / 450
body/sm                  13px / 20px / 450
caption                  12px / 16px / 600
```

## Spacing

```text
space/2                  2px
space/4                  4px
space/8                  8px
space/12                 12px
space/16                 16px
space/20                 20px
space/24                 24px
space/32                 32px
space/40                 40px
space/48                 48px
space/64                 64px
space/80                 80px
space/96                 96px
```

## Radius

Cards stay restrained. Movie posters should look like media, not app bubbles.

```text
radius/control           8px
radius/card              8px
radius/panel             10px
radius/pill              999px
```

## Focus

TV focus must be visible from a couch.

```text
focus/ring-width         3px
focus/ring-color         accent/focus
focus/glow               0 0 0 7px rgba(79, 230, 200, 0.18)
focus/lift               translateY(-6px) scale(1.035)
```

## Motion

Motion should communicate focus, hierarchy, and state. It should not slow browsing.

```text
motion/fast              120ms
motion/base              180ms
motion/slow              260ms
easing/standard          cubic-bezier(0.2, 0.8, 0.2, 1)
easing/enter             cubic-bezier(0.16, 1, 0.3, 1)
```

